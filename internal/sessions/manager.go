package sessions

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/NielsdaWheelz/skidbladnir/internal/catalog"
	"github.com/NielsdaWheelz/skidbladnir/internal/herdr"
	"github.com/NielsdaWheelz/skidbladnir/internal/profile"
	"github.com/NielsdaWheelz/skidbladnir/internal/workdir"
)

// A dwarf's name is a herdr agent name ([a-z][a-z0-9_-]{0,31}) with room for
// the numbered variant that a name collision takes.
var dwarfNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,23}$`)

type Manager struct {
	herdr         *herdr.Client
	workdir       *workdir.Service
	catalogue     catalog.Catalogue
	profiles      []profile.Profile
	profilesByKey map[profile.Key]profile.Profile
	machineHandle string
	mutations     sync.Mutex
}

func New(config Config) (*Manager, error) {
	if config.Herdr == nil || config.Workdir == nil || config.MachineHandle == "" {
		return nil, errors.New("terminal discovery is not configured")
	}
	characters, err := catalog.Load(config.CataloguePath)
	if err != nil {
		return nil, err
	}
	names := make(map[string]bool)
	for _, character := range characters.Characters() {
		name := agentName(character)
		if !dwarfNamePattern.MatchString(name) || names[name] {
			return nil, fmt.Errorf("character %s has no unique herdr agent name", character.Key)
		}
		names[name] = true
	}
	profiles, err := profile.Validate(config.Profiles)
	if err != nil {
		return nil, err
	}
	byKey := make(map[profile.Key]profile.Profile, len(profiles))
	for _, launch := range profiles {
		byKey[launch.Key] = launch
	}
	return &Manager{herdr: config.Herdr, workdir: config.Workdir, catalogue: characters,
		profiles: profiles, profilesByKey: byKey, machineHandle: config.MachineHandle}, nil
}

func (manager *Manager) Profiles() []profile.Profile {
	return profile.Clone(manager.profiles)
}

func (manager *Manager) List(ctx context.Context) (Inventory, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	workspaces, err := manager.workspaces(ctx)
	if err != nil {
		return Inventory{}, err
	}
	panes, err := manager.panes(ctx)
	if err != nil {
		return Inventory{}, err
	}
	inventory := Inventory{Workspaces: make([]Workspace, 0, len(workspaces)), Terminals: make([]Terminal, 0, len(panes))}
	addressable := make(map[string]bool, len(workspaces))
	for _, workspace := range workspaces {
		if !validWorkspaceLabel(workspace.Label) {
			inventory.UnaddressableWorkspaces++
			continue
		}
		addressable[workspace.ID] = true
		inventory.Workspaces = append(inventory.Workspaces, Workspace{Target: WorkspaceTarget{WorkspaceID: workspace.ID}, Label: workspace.Label})
	}
	for _, pane := range panes {
		if !addressable[pane.WorkspaceID] {
			inventory.UnaddressableTerminals++
			continue
		}
		inventory.Terminals = append(inventory.Terminals, manager.project(ctx, pane))
	}
	inventory.ObservedAt = time.Now().UTC()
	inventory.Partial = inventory.UnaddressableWorkspaces > 0 || inventory.UnaddressableTerminals > 0
	return inventory, nil
}

// ResolveTerminal re-reads which pane hosts target's terminal now and returns
// that pane's id for one terminal write.
func (manager *Manager) ResolveTerminal(ctx context.Context, target TerminalTarget) (string, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	pane, err := manager.resolvePane(ctx, target)
	return pane.ID, err
}

func (manager *Manager) resolvePane(ctx context.Context, target TerminalTarget) (paneInfo, error) {
	panes, err := manager.panes(ctx)
	if err != nil {
		return paneInfo{}, err
	}
	for _, pane := range panes {
		if pane.TerminalID == target.TerminalID {
			return pane, nil
		}
	}
	return paneInfo{}, &Error{Code: ErrorTerminalNotFound, Message: "The selected terminal no longer exists.", Dispatch: "not_sent"}
}

func (manager *Manager) observe(ctx context.Context, target TerminalTarget) (Terminal, error) {
	pane, err := manager.resolvePane(ctx, target)
	if err != nil {
		return Terminal{}, err
	}
	return manager.project(ctx, pane), nil
}

func (manager *Manager) project(ctx context.Context, pane paneInfo) Terminal {
	characters := manager.catalogue.Characters()
	seed := sha256.Sum256([]byte(manager.machineHandle + "\x00" + pane.TerminalID))
	terminal := Terminal{
		Target:          TerminalTarget{TerminalID: pane.TerminalID},
		WorkspaceTarget: WorkspaceTarget{WorkspaceID: pane.WorkspaceID},
		Character:       characters[binary.BigEndian.Uint64(seed[:8])%uint64(len(characters))],
	}
	if pane.Label != nil {
		terminal.NativeLabel = *pane.Label
		if pane.Tokens["skid_named"] == "1" && validateName(*pane.Label) == nil {
			terminal.Name = *pane.Label
		}
	}
	if pane.CWD != nil {
		terminal.CWD = *pane.CWD
	}
	if key := profile.Key(pane.Tokens["skid_launch_profile"]); key != "" {
		if _, configured := manager.profilesByKey[key]; configured {
			terminal.LaunchProfile = key
		}
	}
	terminal.Objective = decodeObjective(pane.Tokens)
	terminal.Agent = manager.observeAgent(ctx, pane)
	return terminal
}

// observeAgent projects herdr's agent in one pane as herdr reports it: the
// status from agent get, and readiness only when agent explain matched the
// idle or blocked screen rule for that same observed state. It parses no
// terminal text.
func (manager *Manager) observeAgent(ctx context.Context, pane paneInfo) *Agent {
	if pane.Agent == nil {
		return nil
	}
	provider, known := providerOf(*pane.Agent)
	if !known {
		return nil
	}
	observed, found, err := manager.agent(ctx, pane.ID)
	if err != nil || !found || observed.TerminalID != pane.TerminalID || observed.Agent == nil || *observed.Agent != *pane.Agent {
		return nil
	}
	agent := &Agent{Target: AgentTarget{Terminal: TerminalTarget{TerminalID: pane.TerminalID}, Name: nameOf(observed)},
		Provider: provider, Status: Status{Source: "herdr"}, Readiness: "unconfirmed"}
	switch observed.Status {
	case "working", "blocked", "idle":
		agent.Status.State = observed.Status
	case "done":
		agent.Status.State = "idle"
	default:
		agent.Status.State, agent.Status.Reason = "unknown", "unrecognized"
	}
	var explanation struct {
		Type    string `json:"type"`
		Explain struct {
			MatchedRule *struct {
				State string `json:"state"`
			} `json:"matched_rule"`
			VisibleIdle            bool    `json:"visible_idle"`
			VisibleBlocker         bool    `json:"visible_blocker"`
			ScreenDetectionSkipped bool    `json:"screen_detection_skipped"`
			FallbackReason         *string `json:"fallback_reason"`
		} `json:"explain"`
	}
	if err := manager.herdr.Call(ctx, "agent.explain", map[string]string{"target": pane.ID}, &explanation); err != nil || explanation.Type != "agent_explain" {
		return agent
	}
	explain := explanation.Explain
	if explain.FallbackReason != nil && *explain.FallbackReason == "default_known_agent_idle_fallback" {
		agent.Status.Reason = "default_idle"
	}
	latest, found, err := manager.agent(ctx, pane.ID)
	if err != nil || !found || !sameAgent(latest, agent.Target) || latest.Status != observed.Status || latest.Seq != observed.Seq {
		agent.Status = Status{State: "unknown", Source: "unavailable", Reason: "observation_failed"}
		return agent
	}
	if explain.ScreenDetectionSkipped || explain.FallbackReason != nil || explain.MatchedRule == nil {
		return agent
	}
	if agent.Status.State == "idle" && explain.VisibleIdle && explain.MatchedRule.State == "idle" {
		agent.Readiness = "ready"
	} else if agent.Status.State == "blocked" && explain.VisibleBlocker && explain.MatchedRule.State == "blocked" {
		agent.Readiness = "blocked"
	}
	return agent
}

func providerOf(agent string) (profile.Provider, bool) {
	switch agent {
	case "codex":
		return profile.ProviderCodex, true
	case "claude":
		return profile.ProviderClaude, true
	default:
		return "", false
	}
}

func nameOf(agent agentInfo) string {
	if agent.Name == nil {
		return ""
	}
	return *agent.Name
}

func sameAgent(agent agentInfo, target AgentTarget) bool {
	return agent.TerminalID == target.Terminal.TerminalID && nameOf(agent) == target.Name
}

// agentName is a dwarf's herdr agent name: the final segment of its catalogue
// key, e.g. haugspori for norse.haugspori.
func agentName(character catalog.Character) string {
	return character.Key[strings.LastIndexByte(character.Key, '.')+1:]
}

func (manager *Manager) resolveWorkspace(ctx context.Context, target WorkspaceTarget) error {
	workspaces, err := manager.workspaces(ctx)
	if err != nil {
		return err
	}
	for _, workspace := range workspaces {
		if workspace.ID == target.WorkspaceID && validWorkspaceLabel(workspace.Label) {
			return nil
		}
	}
	return &Error{Code: ErrorWorkspaceStale, Message: "The selected workspace changed.", Dispatch: "not_sent"}
}

func (manager *Manager) generatedName(ctx context.Context, profile string) (string, error) {
	panes, err := manager.panes(ctx)
	if err != nil {
		return "", err
	}
	used := make(map[string]bool, len(panes))
	for _, pane := range panes {
		if pane.Label != nil {
			used[*pane.Label] = true
		}
	}
	for index := 1; ; index++ {
		name := fmt.Sprintf("skidbladnir-%s-%d", profile, index)
		if !used[name] {
			return name, nil
		}
	}
}
