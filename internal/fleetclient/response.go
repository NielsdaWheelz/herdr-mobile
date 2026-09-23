package fleetclient

import (
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/NielsdaWheelz/skidbladnir/internal/reference"
	"github.com/NielsdaWheelz/skidbladnir/internal/strictjson"
)

type Status struct {
	State  string `json:"state"`
	Source string `json:"source"`
	Reason string `json:"reason,omitempty"`
}
type Methods struct {
	Read      string `json:"read"`
	Send      string `json:"send"`
	Interrupt string `json:"interrupt"`
}
type ProviderSession struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
type Agent struct {
	Ref                  string           `json:"ref"`
	Provider             string           `json:"provider"`
	ProvenRuntimeProfile string           `json:"provenRuntimeProfile,omitempty"`
	ProviderSession      *ProviderSession `json:"providerSession,omitempty"`
	Status               Status           `json:"status"`
	Readiness            string           `json:"readiness"`
	Methods              Methods          `json:"methods"`
}
type Character struct {
	Key         string `json:"key"`
	DisplayName string `json:"displayName"`
}
type Terminal struct {
	Ref           string    `json:"ref"`
	Name          string    `json:"name,omitempty"`
	NativeLabel   string    `json:"nativeLabel,omitempty"`
	Character     Character `json:"character"`
	WorkspaceRef  string    `json:"workspaceRef"`
	CWD           string    `json:"cwd,omitempty"`
	LaunchProfile string    `json:"launchProfile,omitempty"`
	Objective     string    `json:"objective,omitempty"`
	Agent         *Agent    `json:"agent,omitempty"`
}
type Workspace struct {
	Ref   string `json:"ref"`
	Label string `json:"label"`
}
type Profile struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Provider string `json:"provider"`
}
type Peer struct {
	Label                   string      `json:"label"`
	Machine                 string      `json:"machine"`
	OK                      bool        `json:"ok"`
	ObservedAt              string      `json:"observedAt,omitempty"`
	Partial                 bool        `json:"partial,omitempty"`
	UnaddressableTerminals  int         `json:"unaddressableTerminals,omitempty"`
	UnaddressableWorkspaces int         `json:"unaddressableWorkspaces,omitempty"`
	Profiles                []Profile   `json:"profiles,omitempty"`
	Workspaces              []Workspace `json:"workspaces,omitempty"`
	Terminals               []Terminal  `json:"terminals,omitempty"`
	Error                   *Failure    `json:"error,omitempty"`
}
type Inventory struct {
	Partial bool   `json:"partial"`
	Peers   []Peer `json:"peers"`
}
type ObservedTerminal struct {
	Label      string   `json:"label"`
	Machine    string   `json:"machine"`
	ObservedAt string   `json:"observedAt"`
	Terminal   Terminal `json:"terminal"`
	Launch     string   `json:"launch,omitempty"`
	Dispatch   string   `json:"dispatch,omitempty"`
}
type ReadResult struct {
	Text      string `json:"text"`
	Source    string `json:"source"`
	Scope     string `json:"scope"`
	Truncated bool   `json:"truncated"`
	Dispatch  string `json:"dispatch,omitempty"`
}
type WriteResult struct {
	Method   string `json:"method"`
	Outcome  string `json:"outcome"`
	Dispatch string `json:"dispatch"`
}
type StopResult struct {
	Agent    string `json:"agent"`
	Terminal string `json:"terminal"`
	Dispatch string `json:"dispatch"`
}
type KillResult struct {
	Terminal string `json:"terminal"`
	Dispatch string `json:"dispatch"`
}

type hostInventory struct {
	Machine struct {
		Handle   string `json:"handle"`
		Platform string `json:"platform"`
	} `json:"machine"`
	ObservedAt              string      `json:"observedAt"`
	Partial                 bool        `json:"partial"`
	UnaddressableTerminals  int         `json:"unaddressableTerminals"`
	UnaddressableWorkspaces int         `json:"unaddressableWorkspaces"`
	Profiles                []Profile   `json:"profiles"`
	Workspaces              []Workspace `json:"workspaces"`
	Terminals               []Terminal  `json:"terminals"`
}

func validTime(value string) bool { _, err := time.Parse(time.RFC3339Nano, value); return err == nil }
func validRef(encoded, machine, kind string) bool {
	value, err := reference.Decode(encoded)
	return err == nil && value.Machine == machine && value.Kind == kind
}
func validTerminal(value Terminal, machine string) bool {
	if !validRef(value.Ref, machine, "terminal") || !validRef(value.WorkspaceRef, machine, "workspace") || value.Character.Key == "" || value.Character.DisplayName == "" {
		return false
	}
	if value.Agent == nil {
		return true
	}
	a := value.Agent
	if !validRef(a.Ref, machine, "agent") || !slices.Contains([]string{"Codex", "Claude"}, a.Provider) || !slices.Contains([]string{"working", "blocked", "idle", "unknown"}, a.Status.State) || !slices.Contains([]string{"herdr", "unavailable"}, a.Status.Source) || !slices.Contains([]string{"", "default_idle", "unrecognized", "observation_failed"}, a.Status.Reason) || !slices.Contains([]string{"ready", "blocked", "unconfirmed"}, a.Readiness) {
		return false
	}
	for _, method := range []string{a.Methods.Read, a.Methods.Send, a.Methods.Interrupt} {
		if method != "terminal" && method != "unavailable" {
			return false
		}
	}
	terminal, _ := reference.Decode(value.Ref)
	agent, _ := reference.Decode(a.Ref)
	return terminal.TerminalID == agent.TerminalID && terminal.IdentityToken == agent.IdentityToken && agent.Agent.Provider == a.Provider
}
func required(encoded []byte, fields ...string) bool {
	var object map[string]json.RawMessage
	if strictjson.Decode(encoded, &object) != nil {
		return false
	}
	for _, field := range fields {
		value, ok := object[field]
		if !ok || string(value) == "null" {
			return false
		}
	}
	return true
}
func decodeInventory(encoded []byte, target peer) (Peer, error) {
	var host *hostInventory
	if strictjson.Decode(encoded, &host) != nil || !required(encoded, "machine", "observedAt", "partial", "unaddressableTerminals", "unaddressableWorkspaces", "profiles", "workspaces", "terminals") || host == nil || host.Machine.Handle != target.Machine || !slices.Contains([]string{"Linux", "Darwin"}, host.Machine.Platform) || !validTime(host.ObservedAt) || host.Profiles == nil || host.Workspaces == nil || host.Terminals == nil || host.UnaddressableTerminals < 0 || host.UnaddressableWorkspaces < 0 {
		return Peer{}, errors.New("invalid inventory")
	}
	for _, p := range host.Profiles {
		if p.Key == "" || p.Label == "" || !slices.Contains([]string{"Codex", "Claude"}, p.Provider) {
			return Peer{}, errors.New("invalid profile")
		}
	}
	for _, w := range host.Workspaces {
		if !validRef(w.Ref, target.Machine, "workspace") || w.Label == "" {
			return Peer{}, errors.New("invalid workspace")
		}
	}
	for _, t := range host.Terminals {
		if !validTerminal(t, target.Machine) {
			return Peer{}, errors.New("invalid terminal")
		}
	}
	return Peer{Label: target.Label, Machine: target.Machine, OK: true, ObservedAt: host.ObservedAt, Partial: host.Partial, UnaddressableTerminals: host.UnaddressableTerminals, UnaddressableWorkspaces: host.UnaddressableWorkspaces, Profiles: host.Profiles, Workspaces: host.Workspaces, Terminals: host.Terminals}, nil
}
func decodeSuccess(operation string, encoded []byte, target peer) (any, error) {
	switch operation {
	case "list":
		return decodeInventory(encoded, target)
	case "info", "start", "shell", "rename", "move":
		var host *struct {
			ObservedAt string   `json:"observedAt"`
			Terminal   Terminal `json:"terminal"`
			Launch     string   `json:"launch,omitempty"`
			Dispatch   string   `json:"dispatch,omitempty"`
		}
		if strictjson.Decode(encoded, &host) != nil || !required(encoded, "observedAt", "terminal") || host == nil || !validTime(host.ObservedAt) || !validTerminal(host.Terminal, target.Machine) {
			return nil, errors.New("invalid terminal response")
		}
		if operation == "info" && (host.Launch != "" || host.Dispatch != "") {
			return nil, errors.New("invalid terminal response")
		}
		if operation != "info" && !required(encoded, "dispatch") {
			return nil, errors.New("missing dispatch")
		}
		if (operation == "start" || operation == "shell") && (!slices.Contains([]string{"submitted", "not_requested"}, host.Launch) || host.Dispatch != "sent") {
			return nil, errors.New("invalid creation")
		}
		if (operation == "rename" || operation == "move") && (host.Launch != "" || host.Dispatch != "sent") {
			return nil, errors.New("invalid mutation")
		}
		return ObservedTerminal{target.Label, target.Machine, host.ObservedAt, host.Terminal, host.Launch, host.Dispatch}, nil
	case "read", "send", "keys", "interrupt", "stop", "kill":
		var fields map[string]json.RawMessage
		if strictjson.Decode(encoded, &fields) != nil || fields == nil {
			return nil, errors.New("invalid response")
		}
		switch operation {
		case "read":
			var value ReadResult
			if strictjson.Decode(encoded, &value) != nil || !required(encoded, "text", "source", "scope", "truncated") || value.Source != "terminal" || !slices.Contains([]string{"visible", "terminal_history"}, value.Scope) {
				return nil, errors.New("invalid read")
			}
		case "send", "keys", "interrupt":
			var value WriteResult
			if strictjson.Decode(encoded, &value) != nil || !required(encoded, "method", "outcome", "dispatch") || value.Method != "terminal" || !slices.Contains([]string{"written", "unknown"}, value.Outcome) || !slices.Contains([]string{"sent", "unknown"}, value.Dispatch) {
				return nil, errors.New("invalid write")
			}
		case "stop":
			var value StopResult
			if strictjson.Decode(encoded, &value) != nil || !required(encoded, "agent", "terminal", "dispatch") || !slices.Contains([]string{"interrupt_sent", "exited"}, value.Agent) || value.Terminal != "closed" || value.Dispatch != "sent" {
				return nil, errors.New("invalid stop")
			}
		case "kill":
			var value KillResult
			if strictjson.Decode(encoded, &value) != nil || !required(encoded, "terminal", "dispatch") || value.Terminal != "closed" || value.Dispatch != "sent" {
				return nil, errors.New("invalid kill")
			}
		}
		fields["label"], _ = json.Marshal(target.Label)
		fields["machine"], _ = json.Marshal(target.Machine)
		return fields, nil
	default:
		return nil, errors.New("unknown operation")
	}
}

func (value Peer) MarshalJSON() ([]byte, error) {
	if !value.OK {
		failure := struct{ Code, Message string }{"unavailable", "Peer unavailable."}
		if value.Error != nil {
			failure.Code = value.Error.Code
			failure.Message = value.Error.Message
			if failure.Message == "" {
				failure.Message = "Peer unavailable."
			}
		}
		return json.Marshal(struct {
			Label   string `json:"label"`
			Machine string `json:"machine"`
			OK      bool   `json:"ok"`
			Error   any    `json:"error"`
		}{value.Label, value.Machine, false, map[string]string{"code": failure.Code, "message": failure.Message}})
	}
	return json.Marshal(struct {
		Label                   string      `json:"label"`
		Machine                 string      `json:"machine"`
		OK                      bool        `json:"ok"`
		ObservedAt              string      `json:"observedAt"`
		Partial                 bool        `json:"partial"`
		UnaddressableTerminals  int         `json:"unaddressableTerminals"`
		UnaddressableWorkspaces int         `json:"unaddressableWorkspaces"`
		Profiles                []Profile   `json:"profiles"`
		Workspaces              []Workspace `json:"workspaces"`
		Terminals               []Terminal  `json:"terminals"`
	}{value.Label, value.Machine, true, value.ObservedAt, value.Partial, value.UnaddressableTerminals, value.UnaddressableWorkspaces, value.Profiles, value.Workspaces, value.Terminals})
}
