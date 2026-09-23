package sessions

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/NielsdaWheelz/skidbladnir/internal/agentruntime"
	"github.com/NielsdaWheelz/skidbladnir/internal/catalog"
	"github.com/NielsdaWheelz/skidbladnir/internal/herdr"
	"github.com/NielsdaWheelz/skidbladnir/internal/process"
	"github.com/NielsdaWheelz/skidbladnir/internal/workdir"
)

type Manager struct {
	herdr         *herdr.Client
	workdir       *workdir.Service
	catalogue     catalog.Catalogue
	profiles      []agentruntime.Profile
	profilesByKey map[agentruntime.ProfileKey]agentruntime.Profile
	machineHandle string
	fingerprint   func(process.Observation) (string, error)
	mutations     sync.Mutex
}

func New(config Config) (*Manager, error) {
	if config.Herdr == nil || config.Workdir == nil || config.MachineHandle == "" || config.Fingerprint == nil {
		return nil, errors.New("terminal discovery is not configured")
	}
	characters, err := catalog.Load(config.CataloguePath)
	if err != nil {
		return nil, err
	}
	profiles, err := agentruntime.ValidateProfiles(config.Profiles)
	if err != nil {
		return nil, err
	}
	byKey := make(map[agentruntime.ProfileKey]agentruntime.Profile, len(profiles))
	for _, profile := range profiles {
		byKey[profile.Key] = profile
	}
	return &Manager{herdr: config.Herdr, workdir: config.Workdir, catalogue: characters,
		profiles: profiles, profilesByKey: byKey, machineHandle: config.MachineHandle,
		fingerprint: config.Fingerprint}, nil
}

func (manager *Manager) Profiles() []agentruntime.Profile {
	return agentruntime.CloneProfiles(manager.profiles)
}

func (manager *Manager) List(ctx context.Context) (Inventory, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	return manager.list(ctx)
}

func (manager *Manager) list(ctx context.Context) (Inventory, error) {
	workspaces, err := manager.workspaces(ctx)
	if err != nil {
		return Inventory{}, err
	}
	panes, err := manager.panes(ctx)
	if err != nil {
		return Inventory{}, err
	}
	inventory := Inventory{Workspaces: make([]Workspace, 0, len(workspaces)), Terminals: make([]Terminal, 0, len(panes))}
	workspaceTargets := make(map[string]WorkspaceTarget, len(workspaces))
	for _, workspace := range workspaces {
		if !validWorkspaceLabel(workspace.Label) {
			inventory.UnaddressableWorkspaces++
			continue
		}
		target, err := manager.claimWorkspace(ctx, workspace)
		if err != nil {
			inventory.UnaddressableWorkspaces++
			continue
		}
		workspaceTargets[workspace.ID] = target
		inventory.Workspaces = append(inventory.Workspaces, Workspace{Target: target, Label: workspace.Label})
	}
	for _, pane := range panes {
		workspaceTarget, addressable := workspaceTargets[pane.WorkspaceID]
		if !addressable {
			inventory.UnaddressableTerminals++
			continue
		}
		target, err := manager.claimPane(ctx, pane)
		if err != nil {
			inventory.UnaddressableTerminals++
			continue
		}
		// Metadata returned by pane.list preceded the token claim. Its other
		// fields may have changed, so project one current pane read.
		current, err := manager.pane(ctx, pane.ID)
		if err != nil || current.TerminalID != target.TerminalID || current.WorkspaceID != workspaceTarget.WorkspaceID || current.Tokens[lifetimeKey] != target.IdentityToken {
			inventory.UnaddressableTerminals++
			continue
		}
		inventory.Terminals = append(inventory.Terminals, manager.project(ctx, current, target, workspaceTarget))
	}
	inventory.ObservedAt = time.Now().UTC()
	inventory.Partial = inventory.UnaddressableWorkspaces > 0 || inventory.UnaddressableTerminals > 0
	return inventory, nil
}

func (manager *Manager) Info(ctx context.Context, target TerminalTarget) (ObservedTerminal, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	resolved, err := manager.resolveTerminal(ctx, target)
	if err != nil {
		return ObservedTerminal{}, err
	}
	return ObservedTerminal{ObservedAt: time.Now().UTC(), Terminal: resolved.Terminal}, nil
}

func (manager *Manager) ResolveTerminal(ctx context.Context, target TerminalTarget) (ResolvedTerminal, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	return manager.resolveTerminal(ctx, target)
}

func (manager *Manager) resolveTerminal(ctx context.Context, target TerminalTarget) (ResolvedTerminal, error) {
	if target.TerminalID == "" || !validLifetime(target.IdentityToken) {
		return ResolvedTerminal{}, &Error{Code: ErrorTerminalStale, Message: "The selected terminal changed.", Dispatch: "not_sent"}
	}
	panes, err := manager.panes(ctx)
	if err != nil {
		return ResolvedTerminal{}, err
	}
	var found *paneInfo
	for index := range panes {
		if panes[index].TerminalID == target.TerminalID {
			found = &panes[index]
			break
		}
	}
	if found == nil {
		return ResolvedTerminal{}, &Error{Code: ErrorTerminalNotFound, Message: "The selected terminal no longer exists.", Dispatch: "not_sent"}
	}
	if found.Tokens[lifetimeKey] != target.IdentityToken {
		return ResolvedTerminal{}, &Error{Code: ErrorTerminalStale, Message: "The selected terminal changed.", Dispatch: "not_sent"}
	}
	workspaces, err := manager.workspaces(ctx)
	if err != nil {
		return ResolvedTerminal{}, err
	}
	for _, workspace := range workspaces {
		if workspace.ID == found.WorkspaceID && validWorkspaceLabel(workspace.Label) && validLifetime(workspace.Tokens[lifetimeKey]) {
			workspaceTarget := WorkspaceTarget{WorkspaceID: workspace.ID, IdentityToken: workspace.Tokens[lifetimeKey]}
			current, err := manager.pane(ctx, found.ID)
			if err != nil || current.TerminalID != target.TerminalID || current.WorkspaceID != workspace.ID || current.Tokens[lifetimeKey] != target.IdentityToken {
				return ResolvedTerminal{}, &Error{Code: ErrorTerminalStale, Message: "The selected terminal changed.", Dispatch: "not_sent"}
			}
			return ResolvedTerminal{Terminal: manager.project(ctx, current, target, workspaceTarget), PaneID: current.ID}, nil
		}
	}
	return ResolvedTerminal{}, &Error{Code: ErrorWorkspaceStale, Message: "The terminal workspace changed.", Dispatch: "not_sent"}
}

func (manager *Manager) ResolveAgent(ctx context.Context, target AgentTarget) (ResolvedAgent, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	resolved, err := manager.resolveTerminal(ctx, target.TerminalTarget)
	if err != nil {
		return ResolvedAgent{}, err
	}
	if resolved.Terminal.Agent == nil || resolved.Terminal.Agent.Target != target {
		return ResolvedAgent{}, &Error{Code: ErrorAgentStale, Message: "The selected agent changed.", Dispatch: "not_sent"}
	}
	observation, err := process.Observe(target.PID)
	if err != nil || observation.StartIdentity != target.StartIdentity {
		return ResolvedAgent{}, &Error{Code: ErrorAgentStale, Message: "The selected agent changed.", Dispatch: "not_sent"}
	}
	fingerprint, err := manager.fingerprint(observation)
	if err != nil || fingerprint != target.CommandFingerprint {
		return ResolvedAgent{}, &Error{Code: ErrorAgentStale, Message: "The selected agent changed.", Dispatch: "not_sent"}
	}
	return ResolvedAgent{Terminal: resolved.Terminal, PaneID: resolved.PaneID, Process: observation}, nil
}

// revalidateStop checks the original worker after an interrupt. A vanished
// worker permits native close only when the pane has returned to its shell;
// an unclassified foreground successor is still a different worker.
func (manager *Manager) revalidateStop(ctx context.Context, target AgentTarget) (string, bool, error) {
	resolved, err := manager.resolveTerminal(ctx, target.TerminalTarget)
	if err != nil {
		return "", false, err
	}
	stale := func() error {
		return &Error{Code: ErrorAgentStale, Message: "The selected agent changed.", Dispatch: "not_sent"}
	}
	if resolved.Terminal.Agent != nil && resolved.Terminal.Agent.Target != target {
		return "", false, stale()
	}
	observed, observationErr := process.Observe(target.PID)
	if observationErr == nil && observed.StartIdentity == target.StartIdentity {
		if resolved.Terminal.Agent == nil {
			return "", false, stale()
		}
		fingerprint, err := manager.fingerprint(observed)
		if err != nil || fingerprint != target.CommandFingerprint {
			return "", false, stale()
		}
		return resolved.PaneID, false, nil
	}
	if observationErr != nil && !errors.Is(observationErr, process.ErrProcessAbsent) {
		return "", false, stale()
	}
	if resolved.Terminal.Agent != nil {
		return "", false, stale()
	}
	info, err := manager.processInfo(ctx, resolved.PaneID)
	if err != nil || info.ShellPID == nil || info.ForegroundPGID == nil || *info.ForegroundPGID == 0 || len(info.Foreground) == 0 {
		return "", false, stale()
	}
	for _, foreground := range info.Foreground {
		if foreground.PID != *info.ShellPID || process.PID(foreground.PID) == target.PID {
			return "", false, stale()
		}
	}
	shell, err := process.Observe(process.PID(*info.ShellPID))
	if err != nil || shell.ProcessGroup != process.PID(*info.ForegroundPGID) || shell.ForegroundProcessGroup != shell.ProcessGroup {
		return "", false, stale()
	}
	if !isShell(shell.ExecutableBase()) {
		return "", false, stale()
	}
	return resolved.PaneID, true, nil
}

func (manager *Manager) project(ctx context.Context, pane paneInfo, target TerminalTarget, workspaceTarget WorkspaceTarget) Terminal {
	characters := manager.catalogue.Characters()
	seed := sha256.Sum256([]byte(manager.machineHandle + "\x00" + pane.TerminalID + "\x00" + target.IdentityToken))
	index := uint64(0)
	for _, value := range seed[:8] {
		index = index<<8 | uint64(value)
	}
	terminal := Terminal{Target: target, WorkspaceTarget: workspaceTarget, Character: characters[index%uint64(len(characters))]}
	if pane.Label != nil {
		terminal.NativeLabel = *pane.Label
		if pane.Tokens["skid_named"] == "1" && validateName(*pane.Label) == nil {
			terminal.Name = *pane.Label
		}
	}
	if pane.CWD != nil {
		terminal.CWD = *pane.CWD
	}
	if profile := agentruntime.ProfileKey(pane.Tokens["skid_launch_profile"]); profile != "" {
		if _, configured := manager.profilesByKey[profile]; configured {
			terminal.LaunchProfile = profile
		}
	}
	terminal.Objective = decodeObjective(pane.Tokens)
	terminal.Agent = manager.observeAgent(ctx, pane, target)
	return terminal
}

func (manager *Manager) observeAgent(ctx context.Context, pane paneInfo, terminal TerminalTarget) *Agent {
	if pane.Agent == nil {
		return nil
	}
	var provider agentruntime.Provider
	switch strings.ToLower(*pane.Agent) {
	case "codex":
		provider = agentruntime.ProviderCodex
	case "claude":
		provider = agentruntime.ProviderClaude
	default:
		return nil
	}
	info, err := manager.processInfo(ctx, pane.ID)
	if err != nil || info.ForegroundPGID == nil || *info.ForegroundPGID == 0 {
		return nil
	}
	var candidates []process.Observation
	var foreground []process.Observation
	for _, listed := range info.Foreground {
		observed, err := process.Observe(process.PID(listed.PID))
		if err != nil || uint32(observed.ProcessGroup) != *info.ForegroundPGID {
			continue
		}
		// The launch shell may exec the worker without changing its PID.
		if info.ShellPID != nil && listed.PID == *info.ShellPID && isShell(observed.ExecutableBase()) {
			continue
		}
		foreground = append(foreground, observed)
		classified, known := agentruntime.ClassifyForeground(manager.profiles, observed)
		if known && classified.Provider == provider ||
			provider == agentruntime.ProviderCodex && observed.ExecutableBase() == "codex" ||
			provider == agentruntime.ProviderClaude && observed.ExecutableBase() == "claude" {
			candidates = append(candidates, observed)
		}
	}
	var candidate process.Observation
	switch len(candidates) {
	case 1:
		candidate = candidates[0]
	case 0:
		// Herdr supplies provider classification. On a zero-profile host, one
		// independently observed non-shell foreground process is the worker.
		if len(foreground) == 1 {
			candidate = foreground[0]
		}
	case 2:
		if provider != agentruntime.ProviderCodex {
			return nil
		}
		for _, wrapper := range candidates {
			if wrapper.ExecutableBase() != "node" {
				continue
			}
			for _, native := range candidates {
				if native.ExecutableBase() == "codex" && native.ParentPID == wrapper.PID {
					candidate = wrapper
				}
			}
		}
	default:
		return nil
	}
	if candidate.PID <= 0 {
		return nil
	}
	fingerprint, err := manager.fingerprint(candidate)
	if err != nil || fingerprint == "" {
		return nil
	}
	target := AgentTarget{TerminalTarget: terminal, PID: candidate.PID, StartIdentity: candidate.StartIdentity,
		CommandFingerprint: fingerprint, Provider: provider}
	agent := &Agent{Target: target, Provider: provider, Status: agentruntime.Status{State: "unknown", Source: "unavailable", Reason: "unrecognized"},
		Readiness: "unconfirmed", Methods: agentruntime.Methods{Read: "terminal", Send: "terminal", Interrupt: "terminal"}}
	if pane.Tokens[agentruntime.RegistrationToken] != "" {
		if registration, ok := agentruntime.ProjectRegistration(manager.profiles, candidate, provider, pane.Tokens); ok {
			agent.ProvenRuntimeProfile = registration.Profile
			agent.ProviderSession = registration.ProviderSession
		}
	}
	return agent
}

func isShell(executable string) bool {
	name := strings.ToLower(strings.TrimSuffix(strings.TrimLeft(executable, "-"), ".exe"))
	switch name {
	case "sh", "bash", "dash", "ash", "zsh", "fish", "ksh", "mksh", "csh", "tcsh",
		"elvish", "xonsh", "nu", "pwsh", "powershell", "cmd":
		return true
	default:
		return false
	}
}

func (manager *Manager) resolveWorkspace(ctx context.Context, target WorkspaceTarget) (workspaceInfo, error) {
	if target.WorkspaceID == "" || !validLifetime(target.IdentityToken) {
		return workspaceInfo{}, &Error{Code: ErrorWorkspaceStale, Message: "The selected workspace changed.", Dispatch: "not_sent"}
	}
	workspaces, err := manager.workspaces(ctx)
	if err != nil {
		return workspaceInfo{}, err
	}
	for _, workspace := range workspaces {
		if workspace.ID == target.WorkspaceID && validWorkspaceLabel(workspace.Label) && workspace.Tokens[lifetimeKey] == target.IdentityToken {
			return workspace, nil
		}
	}
	return workspaceInfo{}, &Error{Code: ErrorWorkspaceStale, Message: "The selected workspace changed.", Dispatch: "not_sent"}
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
