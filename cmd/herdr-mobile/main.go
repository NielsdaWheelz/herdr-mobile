package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/NielsdaWheelz/herdr-mobile/internal/auth"
	"github.com/NielsdaWheelz/herdr-mobile/internal/gateway"
	"github.com/NielsdaWheelz/herdr-mobile/internal/herdr"
	"github.com/NielsdaWheelz/herdr-mobile/internal/hostconfig"
	"github.com/NielsdaWheelz/herdr-mobile/internal/logging"
	"github.com/NielsdaWheelz/herdr-mobile/internal/machine"
	"github.com/NielsdaWheelz/herdr-mobile/internal/pairing"
	"github.com/NielsdaWheelz/herdr-mobile/internal/platform"
	"github.com/NielsdaWheelz/herdr-mobile/internal/pressure"
	"github.com/NielsdaWheelz/herdr-mobile/internal/sessions"
	"github.com/NielsdaWheelz/herdr-mobile/internal/strictjson"
	"github.com/NielsdaWheelz/herdr-mobile/internal/workdir"
)

const (
	exitFailure          = 1
	exitUsage            = 64
	pairingInviteOrigin  = "http://127.0.0.1:7342"
	pairingInviteTimeout = 5 * time.Second
)

const usage = "usage: herdr-mobile {version|validate-host-config|gateway|machine init|bearer mint|pairing-invite create}\n"

var (
	releaseVersion = "dev"
	releaseSHA     = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(arguments []string, stdout, stderr io.Writer) int {
	if len(arguments) == 0 {
		_, _ = io.WriteString(stderr, usage) // justify-ignore-error: a broken CLI output stream cannot be recovered.
		return exitUsage
	}
	if arguments[0] == "version" {
		if len(arguments) != 1 {
			_, _ = io.WriteString(stderr, "usage: herdr-mobile version\n") // justify-ignore-error: a broken CLI output stream cannot be recovered.
			return exitUsage
		}
		if _, err := fmt.Fprintf(stdout, "%s %s\n", releaseVersion, releaseSHA); err != nil {
			_, _ = io.WriteString(stderr, "write version: output failed\n") // justify-ignore-error: both CLI output streams are unavailable.
			return exitFailure
		}
		return 0
	}
	if arguments[0] == "validate-host-config" {
		// Validates configuration only: no herdr binary, socket or home is read.
		path, found := "", false
		if len(arguments) == 2 {
			path, found = strings.CutPrefix(arguments[1], "--host-config=")
		}
		if !found || !filepath.IsAbs(path) {
			_, _ = io.WriteString(stderr, "usage: herdr-mobile validate-host-config --host-config=ABSOLUTE_PATH\n") // justify-ignore-error: a broken CLI output stream cannot be recovered.
			return exitUsage
		}
		if _, err := hostconfig.Load(path, platform.Current().Kind); err != nil {
			_, _ = fmt.Fprintf(stderr, "validate-host-config: %s\n", strings.ReplaceAll(err.Error(), "\n", "; ")) // justify-ignore-error: a broken CLI output stream cannot be recovered.
			return exitFailure
		}
		if _, err := io.WriteString(stdout, "host config valid\n"); err != nil {
			_, _ = io.WriteString(stderr, "write validation: output failed\n") // justify-ignore-error: both CLI output streams are unavailable.
			return exitFailure
		}
		return 0
	}
	home, err := os.UserHomeDir()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "resolve service home: %v\n", err) // justify-ignore-error: a broken CLI output stream cannot be recovered.
		return exitFailure
	}
	switch arguments[0] {
	case "pairing-invite":
		if len(arguments) < 2 || arguments[1] != "create" {
			_, _ = io.WriteString(stderr, "usage: herdr-mobile pairing-invite create [--bearer-file=PATH] [--machine-handle-file=PATH]\n") // justify-ignore-error: a broken CLI output stream cannot be recovered.
			return exitUsage
		}
		flags := flag.NewFlagSet("pairing-invite create", flag.ContinueOnError)
		flags.SetOutput(stderr)
		bearerPath := flags.String("bearer-file", filepath.Join(home, ".config", "herdr-mobile", "bearer"), "bearer file")
		machineHandlePath := flags.String("machine-handle-file", filepath.Join(home, ".config", "herdr-mobile", "machine-handle"), "machine handle file")
		if err := flags.Parse(arguments[2:]); err != nil || flags.NArg() != 0 {
			return exitUsage
		}
		handle, err := machine.Load(*machineHandlePath)
		if err != nil {
			_, _ = io.WriteString(stderr, "pairing-invite create failed\n") // justify-ignore-error: a broken CLI output stream cannot be recovered.
			return exitFailure
		}
		credential, err := (auth.FileVerifier{Path: *bearerPath}).Read()
		if err != nil {
			_, _ = io.WriteString(stderr, "pairing-invite create failed\n") // justify-ignore-error: a broken CLI output stream cannot be recovered.
			return exitFailure
		}
		client := &http.Client{Timeout: pairingInviteTimeout}
		invite, err := requestPairingInvitation(context.Background(), client, pairingInviteOrigin, handle, platform.Current().Kind, credential)
		if err != nil {
			_, _ = io.WriteString(stderr, "pairing-invite create failed\n") // justify-ignore-error: a broken CLI output stream cannot be recovered.
			return exitFailure
		}
		if err := json.NewEncoder(stdout).Encode(invite); err != nil {
			_, _ = io.WriteString(stderr, "write pairing invitation: output failed\n") // justify-ignore-error: both CLI output streams are unavailable.
			return exitFailure
		}
		return 0
	case "machine":
		if len(arguments) == 1 || arguments[1] != "init" {
			_, _ = io.WriteString(stderr, "usage: herdr-mobile machine init [--file=PATH]\n") // justify-ignore-error: a broken CLI output stream cannot be recovered.
			return exitUsage
		}
		flags := flag.NewFlagSet("machine init", flag.ContinueOnError)
		flags.SetOutput(stderr)
		path := flags.String("file", filepath.Join(home, ".config", "herdr-mobile", "machine-handle"), "machine handle file")
		if err := flags.Parse(arguments[2:]); err != nil || flags.NArg() != 0 {
			return exitUsage
		}
		handle, err := machine.Init(*path)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "initialize machine: %v\n", err) // justify-ignore-error: a broken CLI output stream cannot be recovered.
			return exitFailure
		}
		if _, err := fmt.Fprintln(stdout, handle.String()); err != nil {
			_, _ = io.WriteString(stderr, "write machine handle: output failed\n") // justify-ignore-error: both CLI output streams are unavailable.
			return exitFailure
		}
		return 0
	case "bearer":
		if len(arguments) == 1 || arguments[1] != "mint" {
			_, _ = io.WriteString(stderr, "usage: herdr-mobile bearer mint [--file=PATH]\n") // justify-ignore-error: a broken CLI output stream cannot be recovered.
			return exitUsage
		}
		flags := flag.NewFlagSet("bearer mint", flag.ContinueOnError)
		flags.SetOutput(stderr)
		path := flags.String("file", filepath.Join(home, ".config", "herdr-mobile", "bearer"), "bearer file")
		if err := flags.Parse(arguments[2:]); err != nil || flags.NArg() != 0 {
			return exitUsage
		}
		bearer, err := auth.Mint(auth.MintOptions{Path: *path})
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "mint bearer: %v\n", err) // justify-ignore-error: a broken CLI output stream cannot be recovered.
			return exitFailure
		}
		if _, err := fmt.Fprintln(stdout, bearer); err != nil {
			_, _ = io.WriteString(stderr, "write minted bearer: output failed\n") // justify-ignore-error: both CLI output streams are unavailable.
			return exitFailure
		}
		return 0
	case "gateway":
		flags := flag.NewFlagSet("gateway", flag.ContinueOnError)
		flags.SetOutput(stderr)
		listen := flags.String("listen", "127.0.0.1:7342", "numeric loopback listen address")
		bearerPath := flags.String("bearer-file", filepath.Join(home, ".config", "herdr-mobile", "bearer"), "bearer file")
		machineHandlePath := flags.String("machine-handle-file", "", "required machine handle file")
		hostConfigPath := flags.String("host-config", "", "required host config")
		cataloguePath := flags.String("catalogue-path", filepath.Join(home, ".local", "share", "herdr-mobile", "characters.json"), "Dvergatal catalogue")
		if err := flags.Parse(arguments[1:]); err != nil || flags.NArg() != 0 {
			return exitUsage
		}
		if *machineHandlePath == "" || *hostConfigPath == "" {
			_, _ = io.WriteString(stderr, "usage: herdr-mobile gateway --machine-handle-file=PATH --host-config=PATH [options]\n") // justify-ignore-error: a broken CLI output stream cannot be recovered.
			return exitUsage
		}
		if err := serveGateway(*listen, *bearerPath, *machineHandlePath, *hostConfigPath, *cataloguePath, home, stdout); err != nil {
			_, _ = fmt.Fprintf(stderr, "gateway: %v\n", err) // justify-ignore-error: a broken CLI output stream cannot be recovered.
			return exitFailure
		}
		return 0
	default:
		_, _ = io.WriteString(stderr, usage) // justify-ignore-error: a broken CLI output stream cannot be recovered.
		return exitUsage
	}
}

func serveGateway(listen, bearerPath, machineHandlePath, hostConfigPath, cataloguePath, home string, logOutput io.Writer) error {
	handle, err := machine.Load(machineHandlePath)
	if err != nil {
		return fmt.Errorf("load machine handle: %w", err)
	}
	descriptor := platform.Current()
	host, err := hostconfig.Load(hostConfigPath, descriptor.Kind)
	if err != nil {
		return fmt.Errorf("validate host configuration: %w", err)
	}
	client, err := herdr.New(context.Background(), host.Herdr.Path, host.Herdr.SocketPath, host.Herdr.TestedVersion)
	if err != nil {
		return fmt.Errorf("verify configured herdr: %w", err)
	}
	workingDirectories, err := workdir.New(home)
	if err != nil {
		return fmt.Errorf("initialize working directories: %w", err)
	}
	manager, err := sessions.New(sessions.Config{
		Herdr:         client,
		Workdir:       workingDirectories,
		CataloguePath: cataloguePath,
		Profiles:      host.Profiles,
		MachineHandle: handle.String(),
	})
	if err != nil {
		return fmt.Errorf("initialize terminal sessions: %w", err)
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	monitor := pressure.NewMonitor()
	ctx, stopMonitor := context.WithCancel(context.Background())
	defer stopMonitor()
	go monitor.Run(ctx)
	handler := gateway.New(gateway.Config{
		Herdr:    client,
		Sessions: manager,
		Workdir:  workingDirectories,
		Pressure: monitor,
		Bearer:   auth.FileVerifier{Path: bearerPath},
		Pairing:  pairing.NewSlot(),
		Logger:   logging.New(logOutput),
		Machine:  handle,
		Platform: descriptor,
	})
	return gateway.ListenAndServe(signals, listen, handler)
}

type pairingInvitation struct {
	PairingInviteToken string             `json:"pairingInviteToken"`
	ExpiresAt          string             `json:"expiresAt"`
	Machine            pairingMachineWire `json:"machine"`
}

type pairingMachineWire struct {
	Handle   string        `json:"handle"`
	Platform platform.Kind `json:"platform"`
}

func requestPairingInvitation(ctx context.Context, client *http.Client, origin string, handle machine.Handle, expectedPlatform platform.Kind, credential auth.Credential) (pairingInvitation, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, origin+"/v1/pairing-invites", http.NoBody)
	if err != nil {
		return pairingInvitation{}, errors.New("create pairing invitation request")
	}
	request.Header.Set("Authorization", "Bearer "+credential.CanonicalBearer())
	request.Header.Set("Herdr-Mobile-Machine", handle.String())
	response, err := client.Do(request)
	if err != nil {
		return pairingInvitation{}, errors.New("request pairing invitation")
	}
	mediaType, _, mediaErr := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if response.StatusCode != http.StatusCreated || mediaErr != nil || mediaType != "application/json" {
		_ = response.Body.Close() // The rejection is primary; cleanup cannot make the response usable.
		return pairingInvitation{}, errors.New("pairing invitation was rejected")
	}
	encoded, err := io.ReadAll(io.LimitReader(response.Body, gateway.MaximumBodyBytes+1))
	if err != nil || int64(len(encoded)) > gateway.MaximumBodyBytes {
		_ = response.Body.Close() // The read failure is primary; cleanup cannot recover the response.
		return pairingInvitation{}, errors.New("read pairing invitation")
	}
	if err := response.Body.Close(); err != nil {
		return pairingInvitation{}, errors.New("close pairing invitation response")
	}
	var invite pairingInvitation
	if err := strictjson.Decode(encoded, &invite); err != nil {
		return pairingInvitation{}, errors.New("decode pairing invitation")
	}
	parsedHandle, err := machine.Parse(invite.Machine.Handle)
	if err != nil || parsedHandle != handle {
		return pairingInvitation{}, errors.New("pairing invitation machine mismatch")
	}
	if invite.Machine.Platform != expectedPlatform {
		return pairingInvitation{}, errors.New("pairing invitation platform mismatch")
	}
	if _, err := pairing.ParseToken(invite.PairingInviteToken); err != nil {
		return pairingInvitation{}, errors.New("pairing invitation token is invalid")
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, invite.ExpiresAt)
	now := time.Now().UTC()
	if err != nil || !expiresAt.After(now) || expiresAt.After(now.Add(5*time.Minute)) {
		return pairingInvitation{}, errors.New("pairing invitation expiry is invalid")
	}
	return invite, nil
}
