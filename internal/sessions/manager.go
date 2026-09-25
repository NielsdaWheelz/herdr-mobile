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

// A dwarf's name is a herdr agent name ([a-z][a-z0-9_-]{0,31}) with room for,
// and not itself ending in, the numbered variant that a collision takes.
var (
	dwarfNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,23}$`)
	numberedVariant  = regexp.MustCompile(`-[0-9]+$`)
)

// Manager is one host's view of its herdr server. Reads observe herdr without
// the mutation lock; every write re-reads its target and dispatches under it,
// so the gateway's own writes never interleave between a check and its write.
// Another herdr client still can.
type Manager struct {
	herdr         *herdr.Client
	workdir       *workdir.Service
	dwarves       []catalog.Character
	dwarfByName   map[string]catalog.Character
	profiles      []profile.Profile
	profilesByKey map[profile.Key]profile.Profile
	machineHandle string
	mutations     sync.Mutex
}

func New(config Config) (*Manager, error) {
	if config.Herdr == nil || config.Workdir == nil || config.MachineHandle == "" {
		return nil, errors.New("terminal discovery is not configured")
	}
	catalogue, err := catalog.Load(config.CataloguePath)
	if err != nil {
		return nil, err
	}
	dwarves := catalogue.Characters()
	byName := make(map[string]catalog.Character, len(dwarves))
	for _, dwarf := range dwarves {
		name := agentName(dwarf)
		if _, duplicate := byName[name]; duplicate || !dwarfNamePattern.MatchString(name) || numberedVariant.MatchString(name) {
			return nil, fmt.Errorf("character %s has no unique herdr agent name", dwarf.Key)
		}
		byName[name] = dwarf
	}
	profiles, err := profile.Validate(config.Profiles)
	if err != nil {
		return nil, err
	}
	byKey := make(map[profile.Key]profile.Profile, len(profiles))
	for _, launch := range profiles {
		byKey[launch.Key] = launch
	}
	return &Manager{herdr: config.Herdr, workdir: config.Workdir, dwarves: dwarves, dwarfByName: byName,
		profiles: profiles, profilesByKey: byKey, machineHandle: config.MachineHandle}, nil
}

func (manager *Manager) Profiles() []profile.Profile {
	return profile.Clone(manager.profiles)
}

func (manager *Manager) List(ctx context.Context) (Inventory, error) {
	workspaces, err := manager.workspaces(ctx)
	if err != nil {
		return Inventory{}, err
	}
	panes, err := manager.panes(ctx)
	if err != nil {
		return Inventory{}, err
	}
	agents, err := manager.agents(ctx)
	if err != nil {
		return Inventory{}, err
	}
	agentByTerminal := make(map[string]agentInfo, len(agents))
	for _, agent := range agents {
		agentByTerminal[agent.TerminalID] = agent
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
		agent, found := agentByTerminal[pane.TerminalID]
		inventory.Terminals = append(inventory.Terminals, manager.project(ctx, pane, agent, found))
	}
	inventory.ObservedAt = time.Now().UTC()
	inventory.Partial = inventory.UnaddressableWorkspaces > 0 || inventory.UnaddressableTerminals > 0
	return inventory, nil
}

// CheckTerminal re-reads that target's terminal still exists.
func (manager *Manager) CheckTerminal(ctx context.Context, target TerminalTarget) error {
	_, err := manager.resolvePane(ctx, target)
	return err
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
	agent, found, err := manager.agent(ctx, pane.ID)
	if err != nil {
		return Terminal{}, err
	}
	return manager.project(ctx, pane, agent, found), nil
}

// project maps one pane and herdr's agent there, if any, to a terminal. A
// terminal's dwarf is the one its agent's herdr name names, and otherwise the
// one seeded by its terminal id.
func (manager *Manager) project(ctx context.Context, pane paneInfo, agent agentInfo, hasAgent bool) Terminal {
	terminal := Terminal{
		Target:          TerminalTarget{TerminalID: pane.TerminalID},
		WorkspaceTarget: WorkspaceTarget{WorkspaceID: pane.WorkspaceID},
		Character:       manager.dwarves[manager.seed(pane.TerminalID)],
	}
	if hasAgent {
		if dwarf, found := manager.dwarfOf(nameOf(agent)); found {
			terminal.Character = dwarf
		}
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
	if hasAgent {
		terminal.Agent = manager.observeAgent(ctx, pane, agent)
	}
	return terminal
}

// observeAgent projects herdr's agent in one pane as herdr reports it: the
// status from herdr's agent state, and readiness only when agent explain
// matched the idle or blocked screen rule for that same unchanged state. It
// parses no terminal text.
func (manager *Manager) observeAgent(ctx context.Context, pane paneInfo, observed agentInfo) *Agent {
	if pane.Agent == nil || observed.TerminalID != pane.TerminalID || observed.Agent == nil || *observed.Agent != *pane.Agent {
		return nil
	}
	provider, known := providerOf(*pane.Agent)
	if !known {
		return nil
	}
	agent := &Agent{Name: nameOf(observed), Provider: provider, Status: Status{Source: "herdr"}, Readiness: "unconfirmed"}
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
	if agent.Status.State == "idle" && explain.FallbackReason != nil && *explain.FallbackReason == "default_known_agent_idle_fallback" {
		agent.Status.Reason = "default_idle"
	}
	latest, found, err := manager.agent(ctx, pane.ID)
	if err != nil || !found || latest.TerminalID != observed.TerminalID || nameOf(latest) != agent.Name || latest.Status != observed.Status || latest.Seq != observed.Seq {
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

// sameAgent says whether herdr's agent is the target's: the same terminal and
// the same herdr name.
func sameAgent(agent agentInfo, target AgentTarget) bool {
	return agent.TerminalID == target.Terminal.TerminalID && nameOf(agent) == target.Name
}

// seed is the catalogue index a terminal id seeds on this machine.
func (manager *Manager) seed(terminalID string) int {
	digest := sha256.Sum256([]byte(manager.machineHandle + "\x00" + terminalID))
	return int(binary.BigEndian.Uint64(digest[:8]) % uint64(len(manager.dwarves)))
}

// agentName is a dwarf's herdr agent name: the final segment of its catalogue
// key, e.g. haugspori for norse.haugspori.
func agentName(dwarf catalog.Character) string {
	return dwarf.Key[strings.LastIndexByte(dwarf.Key, '.')+1:]
}

// dwarfOf reads a herdr agent name as a dwarf's name or its numbered variant.
func (manager *Manager) dwarfOf(name string) (catalog.Character, bool) {
	if dwarf, found := manager.dwarfByName[name]; found {
		return dwarf, true
	}
	dwarf, found := manager.dwarfByName[numberedVariant.ReplaceAllString(name, "")]
	return dwarf, found
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
