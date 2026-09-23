// Package agentcli parses and renders exact fleet commands.
package agentcli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/NielsdaWheelz/skidbladnir/internal/fleetclient"
	"github.com/NielsdaWheelz/skidbladnir/internal/terminalclient"
	"golang.org/x/term"
)

const usage = `usage: skid COMMAND [options]

bare skid opens the configured local herdr desktop; skid enter attaches through a gateway.
list [--machine HOST] [--json]                   discover terminal and workspace refs
info TARGET                                      inspect a terminal and current agent
start NAME --machine HOST (--profile KEY | --terminal) [--cwd PATH]
           [--objective TEXT] [--workspace-ref REF | --new-workspace LABEL]
shell TARGET                                     create one independent shell in its workspace
rename TARGET NEWNAME                            name a terminal
move TARGET (--workspace-ref REF | --new-workspace LABEL)
read TARGET [--coverage recent|visible] [--max-bytes N]
send TARGET TEXT [--terminal]                     submit once when ready by default
send TARGET --stdin [--terminal]                  submit literal stdin, up to 32 kib
keys TARGET KEY...                                enter escape ctrl-c up down left right tab backspace
interrupt TARGET                                 send one provider interrupt
stop TARGET                                      interrupt, then attempt native terminal close
kill TARGET                                      close the exact terminal
enter TARGET                                     attach; ctrl-] d detaches outside paste

TARGET is an exact case-sensitive name, or --ref REF. --machine HOST qualifies names.
refs are separate for terminal, agent and workspace; preserve returned refs unchanged.
a bare name requires complete, unique inventory. unavailable peers or duplicate names
require --machine or --ref. native herdr labels do not select names.
--config PATH selects the private peer file; default ~/.config/skidbladnir/client.json.
--json emits one {ok,result|error} envelope for noninteractive commands.

start defaults to a new workspace labelled with NAME. move requires a destination.
closing a final pane may close linked workspaces and running terminals. stop first
interrupts its observed worker, then attempts that close. closure may be refused;
never infer that a descendant stopped. a sent input is queue delivery, not provider
processing or task success. unknown dispatch may have taken effect: inspect before
any retry. ordinary send requires recognized readiness; --terminal explicitly
bypasses readiness, while retaining exact worker validation. read defaults to
bounded recent terminal history; --coverage visible reads the current screen.

examples:
  skid list --json
  skid start reviewer --machine arch --profile claude-work --cwd ~/code/project --json
  skid info --ref TERMINAL_REF --json
  skid read --ref AGENT_REF --coverage visible --json
  skid send --ref AGENT_REF --stdin --json < prompt.txt
  skid move --ref TERMINAL_REF --workspace-ref WORKSPACE_REF --json

exit 0 means a reported complete result; 1 means failure, partial inventory or
unknown delivery; 2 means invalid usage. use git/files/ssh to exchange work
products across hosts. skid supplies no completion callback or shared filesystem.
`

type command struct {
	request           fleetclient.Request
	config            string
	json, stdin, help bool
}

func parse(args []string) (command, error) {
	var c command
	var operands []string
	seen := map[string]bool{}
	literal := false
	for index := 0; index < len(args); index++ {
		value := args[index]
		if value == "--" && !literal {
			literal = true
			continue
		}
		if !literal && strings.HasPrefix(value, "--") {
			name, argument, inline := strings.Cut(value, "=")
			if seen[name] {
				return c, errors.New("duplicate option")
			}
			seen[name] = true
			switch name {
			case "--json", "--stdin", "--terminal", "--help":
				if inline {
					return c, errors.New("boolean option has value")
				}
				switch name {
				case "--json":
					c.json = true
				case "--stdin":
					c.stdin = true
				case "--terminal":
					c.request.Mode = "terminal"
				case "--help":
					c.help = true
				}
			case "--config", "--machine", "--ref", "--profile", "--cwd", "--objective", "--coverage", "--max-bytes", "--workspace-ref", "--new-workspace":
				if !inline {
					index++
					if index >= len(args) {
						return c, errors.New("missing option value")
					}
					argument = args[index]
				}
				if argument == "" {
					return c, errors.New("empty option value")
				}
				switch name {
				case "--config":
					c.config = argument
				case "--machine":
					c.request.Machine = argument
				case "--ref":
					c.request.Ref = argument
				case "--profile":
					c.request.Profile = argument
				case "--cwd":
					c.request.CWD = argument
				case "--objective":
					c.request.Objective = argument
				case "--coverage":
					c.request.Coverage = argument
				case "--workspace-ref":
					c.request.DestinationKind = "existing"
					c.request.WorkspaceRef = argument
				case "--new-workspace":
					c.request.DestinationKind = "new"
					c.request.NewWorkspace = argument
				case "--max-bytes":
					n, err := strconv.Atoi(argument)
					if err != nil || n < 1 || n > 32768 {
						return c, errors.New("invalid read limit")
					}
					c.request.MaxBytes = n
				}
			default:
				return c, errors.New("unknown option")
			}
		} else {
			operands = append(operands, value)
		}
	}
	if c.help {
		return c, nil
	}
	if len(operands) == 0 {
		return c, errors.New("missing command")
	}
	c.request.Operation = operands[0]
	operands = operands[1:]
	switch c.request.Operation {
	case "list":
		if len(operands) != 0 {
			return c, errors.New("list takes no target")
		}
	case "start":
		if len(operands) != 1 {
			return c, errors.New("start requires name")
		}
		c.request.Name = operands[0]
		c.request.Kind = fleetclient.LaunchAgent
		if seen["--terminal"] {
			c.request.Kind = fleetclient.LaunchTerminal
			c.request.Mode = ""
		}
		if c.request.CWD == "" {
			c.request.CWD = "~"
		}
	case "info", "shell", "rename", "move", "kill", "enter", "read", "send", "keys", "interrupt", "stop":
		if c.request.Ref == "" {
			if len(operands) == 0 {
				return c, errors.New("missing target")
			}
			c.request.Name = operands[0]
			operands = operands[1:]
		}
		switch c.request.Operation {
		case "rename":
			if len(operands) != 1 {
				return c, errors.New("rename requires new name")
			}
			c.request.NewName = operands[0]
		case "send":
			if c.stdin {
				if len(operands) != 0 {
					return c, errors.New("stdin and text are exclusive")
				}
				c.request.Text = "stdin"
			} else {
				if len(operands) != 1 {
					return c, errors.New("send requires text")
				}
				c.request.Text = operands[0]
			}
		case "keys":
			c.request.Keys = operands
		default:
			if len(operands) != 0 {
				return c, errors.New("extra operand")
			}
		}
	default:
		return c, errors.New("unknown command")
	}
	if c.request.Operation == "start" && seen["--profile"] == seen["--terminal"] {
		return c, errors.New("choose profile or terminal")
	}
	if seen["--terminal"] && c.request.Operation != "start" && c.request.Operation != "send" {
		return c, errors.New("terminal option not supported")
	}
	if seen["--profile"] && c.request.Operation != "start" || seen["--objective"] && c.request.Operation != "start" || seen["--cwd"] && c.request.Operation != "start" {
		return c, errors.New("launch option not supported")
	}
	if seen["--coverage"] && c.request.Operation != "read" || seen["--max-bytes"] && c.request.Operation != "read" {
		return c, errors.New("read option not supported")
	}
	if (seen["--workspace-ref"] || seen["--new-workspace"]) && c.request.Operation != "start" && c.request.Operation != "move" {
		return c, errors.New("destination not supported")
	}
	if seen["--workspace-ref"] && seen["--new-workspace"] {
		return c, errors.New("choose one destination")
	}
	if c.stdin && c.request.Operation != "send" || c.json && c.request.Operation == "enter" {
		return c, errors.New("option not supported")
	}
	if !c.request.Valid() {
		return c, errors.New("invalid command arguments")
	}
	return c, nil
}
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	c, err := parse(args)
	if err != nil {
		for _, arg := range args {
			if arg == "--" {
				break
			}
			if arg == "--json" {
				c.json = true
			}
		}
		if c.json {
			encoded, _ := fleetclient.Failed("invalid_input", "not_sent").Encode("")
			_, _ = stdout.Write(encoded)
		} else {
			_, _ = io.WriteString(stderr, usage)
		}
		return 2
	}
	if c.help {
		if _, err := io.WriteString(stdout, usage); err != nil {
			return 1
		}
		return 0
	}
	if c.request.Operation == "enter" {
		in, okIn := stdin.(*os.File)
		out, okOut := stdout.(*os.File)
		if !okIn || !okOut || !term.IsTerminal(int(in.Fd())) || !term.IsTerminal(int(out.Fd())) {
			_, _ = io.WriteString(stderr, "enter requires stdin and stdout ttys\n")
			return 2
		}
	}
	if c.stdin {
		data, err := io.ReadAll(io.LimitReader(stdin, 32769))
		if err != nil {
			return render(c, fleetclient.Failed("input_unavailable", "not_sent"), stdout, stderr)
		}
		if len(data) > 32768 {
			return render(c, fleetclient.Failed("input_limit", "not_sent"), stdout, stderr)
		}
		c.request.Text = string(data)
		if !c.request.Valid() {
			return render(c, fleetclient.Failed("invalid_input", "not_sent"), stdout, stderr)
		}
	}
	if c.config == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return render(c, fleetclient.Failed("configuration_invalid", "not_sent"), stdout, stderr)
		}
		c.config = filepath.Join(home, ".config", "skidbladnir", "client.json")
	}
	client, err := fleetclient.Open(c.config)
	if err != nil {
		return render(c, fleetclient.Failed("configuration_invalid", "not_sent"), stdout, stderr)
	}
	if c.request.Operation == "enter" {
		_, _ = io.WriteString(stderr, "ctrl-] d detaches; unsupported keys are rejected locally\n")
		if err := terminalclient.Run(ctx, client, c.request, stdin.(*os.File), stdout.(*os.File)); err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}
	return render(c, client.Execute(ctx, c.request), stdout, stderr)
}
func render(c command, result fleetclient.Result, stdout, stderr io.Writer) int {
	if c.json {
		encoded, err := result.Encode(c.request.Operation)
		if err != nil {
			return 1
		}
		if _, err = stdout.Write(encoded); err != nil {
			return 1
		}
		return result.ExitCode(c.request.Operation)
	}
	if !result.OK {
		_, _ = fmt.Fprintf(stderr, "%s: %s (%s)\n", result.Error.Code, result.Error.Message, result.Error.Dispatch)
		if len(result.Partial) > 0 {
			_, _ = fmt.Fprintf(stderr, "partial: %s\n", result.Partial)
		}
		for _, candidate := range result.Candidates {
			_, _ = fmt.Fprintln(stderr, candidate)
		}
		return 1
	}
	switch c.request.Operation {
	case "list":
		value := result.Value.(fleetclient.Inventory)
		table := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
		_, _ = fmt.Fprintln(table, "machine\tworkspace\tterminal\tprovider\tstatus\tdirectory")
		count := 0
		for _, peer := range value.Peers {
			if !peer.OK {
				_, _ = fmt.Fprintf(table, "%s\t\tunavailable\t\t%s\t\n", peer.Label, peer.Error.Code)
				continue
			}
			for _, terminal := range peer.Terminals {
				count++
				name := terminalDisplayName(terminal)
				if terminal.Name == "" && terminal.NativeLabel != "" {
					name += " (herdr label: " + terminal.NativeLabel + ")"
				}
				workspace := ""
				for _, w := range peer.Workspaces {
					if w.Ref == terminal.WorkspaceRef {
						workspace = w.Label
						break
					}
				}
				provider, state := "shell", ""
				if terminal.Agent != nil {
					provider = terminal.Agent.Provider
					state = terminal.Agent.Status.State + "/" + terminal.Agent.Readiness
				}
				_, _ = fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\n", peer.Label, workspace, name, provider, state, terminal.CWD)
			}
		}
		if count == 0 {
			_, _ = fmt.Fprintln(table, "no terminals")
		}
		if table.Flush() != nil {
			return 1
		}
	case "info", "start", "shell", "rename", "move":
		value := result.Value.(fleetclient.ObservedTerminal)
		name := terminalDisplayName(value.Terminal)
		_, _ = fmt.Fprintf(stdout, "terminal: %s\nmachine: %s\nreference: %s\nworkspace reference: %s\n", name, value.Label, value.Terminal.Ref, value.Terminal.WorkspaceRef)
		if value.Terminal.NativeLabel != "" {
			_, _ = fmt.Fprintf(stdout, "herdr label: %s\n", value.Terminal.NativeLabel)
		}
		if value.Terminal.Agent != nil {
			a := value.Terminal.Agent
			_, _ = fmt.Fprintf(stdout, "agent: %s (%s; %s)\nagent reference: %s\n", a.Provider, a.Status.State, a.Readiness, a.Ref)
		}
		if value.Dispatch != "" {
			_, _ = fmt.Fprintf(stdout, "dispatch: %s\n", value.Dispatch)
		}
		if value.Launch != "" {
			_, _ = fmt.Fprintf(stdout, "launch: %s; startup not confirmed\n", value.Launch)
		}
	case "read":
		value := result.Value.(map[string]json.RawMessage)
		var read fleetclient.ReadResult
		encoded, _ := json.Marshal(value)
		if json.Unmarshal(encoded, &read) != nil {
			return 1
		}
		_, _ = fmt.Fprintf(stderr, "source: %s; scope: %s; truncated: %t\n", read.Source, read.Scope, read.Truncated)
		if _, err := io.WriteString(stdout, read.Text); err != nil {
			return 1
		}
	case "send", "keys", "interrupt":
		value := result.Value.(map[string]json.RawMessage)
		var write fleetclient.WriteResult
		encoded, _ := json.Marshal(value)
		if json.Unmarshal(encoded, &write) != nil {
			return 1
		}
		_, _ = fmt.Fprintf(stdout, "%s: %s (%s)\n", write.Method, write.Outcome, write.Dispatch)
	case "stop", "kill":
		value := result.Value.(map[string]json.RawMessage)
		encoded, _ := json.Marshal(value)
		_, _ = fmt.Fprintln(stdout, string(encoded))
	}
	return result.ExitCode(c.request.Operation)
}

func terminalDisplayName(terminal fleetclient.Terminal) string {
	if terminal.Name != "" {
		return terminal.Name
	}
	return "unnamed terminal " + terminal.Ref[len(terminal.Ref)-8:]
}
