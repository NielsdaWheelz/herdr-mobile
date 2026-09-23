package gateway

import (
	"errors"
	"sort"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/NielsdaWheelz/skidbladnir/internal/agentruntime"
	"github.com/NielsdaWheelz/skidbladnir/internal/machine"
	"github.com/NielsdaWheelz/skidbladnir/internal/sessions"
)

type workspaceWire struct {
	Ref   string `json:"ref"`
	Label string `json:"label"`
}

type agentWire struct {
	Ref                  string               `json:"ref"`
	Provider             string               `json:"provider"`
	ProvenRuntimeProfile string               `json:"provenRuntimeProfile,omitempty"`
	ProviderSession      *providerSessionDTO  `json:"providerSession,omitempty"`
	Status               agentruntime.Status  `json:"status"`
	Readiness            string               `json:"readiness"`
	Methods              agentruntime.Methods `json:"methods"`
}

type terminalWire struct {
	Ref           string       `json:"ref"`
	Name          string       `json:"name,omitempty"`
	NativeLabel   string       `json:"nativeLabel,omitempty"`
	Character     characterDTO `json:"character"`
	WorkspaceRef  string       `json:"workspaceRef"`
	CWD           string       `json:"cwd,omitempty"`
	LaunchProfile string       `json:"launchProfile,omitempty"`
	Objective     string       `json:"objective,omitempty"`
	Agent         *agentWire   `json:"agent,omitempty"`
}

type inventoryWire struct {
	Machine                 machineDTO      `json:"machine"`
	ObservedAt              string          `json:"observedAt"`
	Partial                 bool            `json:"partial"`
	UnaddressableTerminals  int             `json:"unaddressableTerminals"`
	UnaddressableWorkspaces int             `json:"unaddressableWorkspaces"`
	Profiles                []profileDTO    `json:"profiles"`
	Workspaces              []workspaceWire `json:"workspaces"`
	Terminals               []terminalWire  `json:"terminals"`
}

func projectionInstant(when time.Time) (string, error) {
	if !sessions.ValidProjectionInstant(when) {
		return "", errors.New("invalid projection instant")
	}
	return when.UTC().Format(time.RFC3339Nano), nil
}

func mapTerminal(handle machine.Handle, terminal sessions.Terminal) (terminalWire, error) {
	ref, err := terminalReference(handle, terminal.Target)
	if err != nil {
		return terminalWire{}, err
	}
	workspaceRef, err := workspaceReference(handle, terminal.WorkspaceTarget)
	if err != nil {
		return terminalWire{}, err
	}
	card := terminalWire{
		Ref: ref, Name: terminal.Name, NativeLabel: terminal.NativeLabel,
		Character:    characterDTO{Key: terminal.Character.Key, DisplayName: terminal.Character.DisplayName},
		WorkspaceRef: workspaceRef, CWD: terminal.CWD,
		LaunchProfile: string(terminal.LaunchProfile), Objective: terminal.Objective,
	}
	if !safeNativeLabel(card.NativeLabel) {
		card.NativeLabel = ""
	}
	if terminal.Agent != nil {
		ref, err := agentReference(handle, terminal.Agent.Target)
		if err != nil {
			return terminalWire{}, err
		}
		agent := terminal.Agent
		card.Agent = &agentWire{
			Ref: ref, Provider: string(agent.Provider),
			ProvenRuntimeProfile: string(agent.ProvenRuntimeProfile),
			Status:               agent.Status, Readiness: agent.Readiness, Methods: agent.Methods,
		}
		if agent.ProviderSession != nil {
			card.Agent.ProviderSession = &providerSessionDTO{ID: agent.ProviderSession.ID(), Name: agent.ProviderSession.Name()}
		}
	}
	return card, nil
}

func safeNativeLabel(value string) bool {
	if value == "" || !utf8.ValidString(value) {
		return false
	}
	for _, symbol := range value {
		if unicode.IsControl(symbol) || symbol == 0x061c || symbol >= 0x200e && symbol <= 0x200f ||
			symbol >= 0x2028 && symbol <= 0x202e || symbol >= 0x2066 && symbol <= 0x2069 {
			return false
		}
	}
	return true
}

func mapInventory(handle machine.Handle, platform machineDTO, inventory sessions.Inventory, profiles []agentruntime.Profile) (inventoryWire, error) {
	observedAt, err := projectionInstant(inventory.ObservedAt)
	if err != nil {
		return inventoryWire{}, err
	}
	card := inventoryWire{
		Machine: platform, ObservedAt: observedAt, Partial: inventory.Partial,
		UnaddressableTerminals:  inventory.UnaddressableTerminals,
		UnaddressableWorkspaces: inventory.UnaddressableWorkspaces,
		Profiles:                make([]profileDTO, 0, len(profiles)),
		Workspaces:              make([]workspaceWire, 0, len(inventory.Workspaces)),
		Terminals:               make([]terminalWire, 0, len(inventory.Terminals)),
	}
	for _, profile := range profiles {
		card.Profiles = append(card.Profiles, profileDTO{Key: string(profile.Key), Label: profile.Label, Provider: string(profile.Provider)})
	}
	for _, workspace := range inventory.Workspaces {
		ref, err := workspaceReference(handle, workspace.Target)
		if err != nil {
			return inventoryWire{}, err
		}
		card.Workspaces = append(card.Workspaces, workspaceWire{Ref: ref, Label: workspace.Label})
	}
	for _, terminal := range inventory.Terminals {
		mapped, err := mapTerminal(handle, terminal)
		if err != nil {
			return inventoryWire{}, err
		}
		card.Terminals = append(card.Terminals, mapped)
	}
	sort.Slice(card.Workspaces, func(i, j int) bool {
		if card.Workspaces[i].Label != card.Workspaces[j].Label {
			return card.Workspaces[i].Label < card.Workspaces[j].Label
		}
		return card.Workspaces[i].Ref < card.Workspaces[j].Ref
	})
	sort.Slice(card.Terminals, func(i, j int) bool {
		if card.Terminals[i].Name != card.Terminals[j].Name {
			return card.Terminals[i].Name < card.Terminals[j].Name
		}
		return card.Terminals[i].Ref < card.Terminals[j].Ref
	})
	return card, nil
}
