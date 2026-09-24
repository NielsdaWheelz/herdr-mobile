package sessions

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/NielsdaWheelz/skidbladnir/internal/catalog"
	"github.com/NielsdaWheelz/skidbladnir/internal/herdr"
	"github.com/NielsdaWheelz/skidbladnir/internal/profile"
	"github.com/NielsdaWheelz/skidbladnir/internal/workdir"
)

// herdr refuses agent.start before writing anything with agent_pane_busy while
// a new pane's shell is still starting, and with agent_name_taken when another
// client took the chosen name. herdr's own cli retries the busy refusal for
// this long, at this cadence; the gateway retries both, never holding the
// mutation lock while it waits.
const (
	agentStartRetryWindow   = 2 * time.Second
	agentStartRetryInterval = 100 * time.Millisecond
)

// Create makes a terminal and, for an agent profile, then starts the agent in
// it without holding the mutation lock across the start's retries.
func (manager *Manager) Create(ctx context.Context, input CreateInput) (Created, error) {
	manager.mutations.Lock()
	terminal, err := manager.create(ctx, input)
	manager.mutations.Unlock()
	if err != nil {
		return Created{}, err
	}
	if input.Kind == LaunchTerminal {
		return Created{ObservedAt: time.Now().UTC(), Terminal: terminal, Launch: "not_requested", Dispatch: "sent"}, nil
	}
	dwarf, err := manager.startAgent(ctx, terminal.Target, manager.profilesByKey[profile.Key(input.Profile)].Provider)
	if err != nil {
		return Created{}, manager.failedLaunch(ctx, terminal, err)
	}
	terminal.Character = dwarf
	return Created{ObservedAt: time.Now().UTC(), Terminal: terminal, Launch: "submitted", Dispatch: "sent"}, nil
}

func (manager *Manager) Shell(ctx context.Context, source TerminalTarget) (Created, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	pane, err := manager.resolvePane(ctx, source)
	if err != nil {
		return Created{}, err
	}
	var cwd string
	if pane.ForegroundCWD != nil {
		cwd = *pane.ForegroundCWD
	} else if pane.CWD != nil {
		cwd = *pane.CWD
	}
	if cwd == "" {
		return Created{}, newSessionError(ErrorWorkingDirectoryUnavailable, "The source directory is unavailable.")
	}
	terminal, err := manager.create(ctx, CreateInput{Kind: LaunchTerminal, CWD: cwd,
		Destination: Destination{Kind: "existing", WorkspaceTarget: WorkspaceTarget{WorkspaceID: pane.WorkspaceID}}})
	if err != nil {
		return Created{}, err
	}
	return Created{ObservedAt: time.Now().UTC(), Terminal: terminal, Launch: "not_requested", Dispatch: "sent"}, nil
}

func (manager *Manager) create(ctx context.Context, input CreateInput) (Terminal, error) {
	candidate, err := manager.workdir.ParseCandidate(input.CWD)
	if err != nil {
		return Terminal{}, mapWorkingDirectoryError(err)
	}
	cwd, err := manager.workdir.ValidateStart(candidate)
	if err != nil {
		return Terminal{}, mapWorkingDirectoryError(err)
	}
	var launch profile.Profile
	namePrefix := "terminal"
	switch input.Kind {
	case LaunchAgent:
		var found bool
		launch, found = manager.profilesByKey[profile.Key(input.Profile)]
		if !found {
			return Terminal{}, newSessionError(ErrorProfileUnknown, "Choose an available profile.")
		}
		namePrefix = string(launch.Key)
	case LaunchTerminal:
		if input.Profile != "" || input.Objective != "" {
			return Terminal{}, newSessionError(ErrorProfileUnknown, "A terminal has no agent profile or objective.")
		}
	default:
		return Terminal{}, newSessionError(ErrorProfileUnknown, "Choose a terminal or configured agent profile.")
	}
	name := input.Name
	if name == "" {
		name, err = manager.generatedName(ctx, namePrefix)
		if err != nil {
			return Terminal{}, err
		}
	}
	if err := validateName(name); err != nil {
		return Terminal{}, err
	}
	if err := validateObjective(input.Objective); err != nil {
		return Terminal{}, err
	}
	destination := input.Destination
	if destination.Kind == "" {
		destination.Kind = "new"
	}
	if destination.Kind == "new" && destination.Label == "" {
		destination.Label = name
	}
	if err := manager.validateDestination(ctx, destination, false); err != nil {
		return Terminal{}, err
	}
	if _, err := manager.workdir.ValidateStart(candidate); err != nil {
		return Terminal{}, mapWorkingDirectoryError(err)
	}
	env := make(map[string]string, len(launch.Environment))
	for _, variable := range launch.Environment {
		env[variable.Name] = variable.Value
	}
	var created struct {
		Type      string        `json:"type"`
		Workspace workspaceInfo `json:"workspace"`
		RootPane  paneInfo      `json:"root_pane"`
	}
	if destination.Kind == "existing" {
		err = manager.herdr.Call(ctx, "tab.create", map[string]any{"workspace_id": destination.WorkspaceTarget.WorkspaceID,
			"cwd": cwd.String(), "focus": false, "env": env}, &created)
		if err == nil && (created.Type != "tab_created" || created.RootPane.ID == "" || created.RootPane.TerminalID == "") {
			err = errors.New("invalid herdr tab creation response")
		}
	} else {
		err = manager.herdr.Call(ctx, "workspace.create", map[string]any{"cwd": cwd.String(), "focus": false,
			"label": destination.Label, "env": env}, &created)
		if err == nil && (created.Type != "workspace_created" || created.Workspace.ID == "" || created.RootPane.ID == "" || created.RootPane.TerminalID == "") {
			err = errors.New("invalid herdr workspace creation response")
		}
	}
	if err != nil {
		return Terminal{}, createFailure(err, "", nil)
	}
	pane := created.RootPane
	target := TerminalTarget{TerminalID: pane.TerminalID}
	// A known create has happened. Every subsequent failure retains that fact.
	if err := manager.renamePane(ctx, pane.ID, name); err != nil {
		return Terminal{}, createFailure(err, "resource_created", nil)
	}
	metadata := map[string]string{"skid_named": "1"}
	if input.Kind == LaunchAgent {
		metadata["skid_launch_profile"] = string(launch.Key)
	}
	if err := manager.reportPane(ctx, pane.ID, metadata); err != nil {
		return Terminal{}, createFailure(err, "resource_created", nil)
	}
	if input.Objective != "" {
		if err := manager.writeObjective(ctx, pane.ID, input.Objective); err != nil {
			return Terminal{}, createFailure(err, "resource_created", nil)
		}
	}
	terminal, err := manager.observe(ctx, target)
	if err != nil {
		return Terminal{}, createFailure(err, "resource_created", nil)
	}
	if terminal.Name != name || input.Objective != "" && terminal.Objective != input.Objective {
		return Terminal{}, createFailure(errors.New("terminal metadata was not retained"), "identified", &terminal)
	}
	return terminal, nil
}

// startAgent names the agent after a dwarf and asks herdr to type the
// provider's bare command at the terminal's shell. The pane's environment
// already carries the profile's account home, and the deployment's shell
// aliases add its permission flags, so skid passes no arguments: a second
// --yolo makes codex refuse to start. It returns the dwarf the agent is named
// after.
func (manager *Manager) startAgent(ctx context.Context, target TerminalTarget, provider profile.Provider) (catalog.Character, error) {
	kind := "codex"
	if provider == profile.ProviderClaude {
		kind = "claude"
	}
	deadline := time.Now().Add(agentStartRetryWindow)
	for {
		dwarf, err := manager.tryStartAgent(ctx, target, kind)
		var upstream *herdr.Error
		if !errors.As(err, &upstream) {
			return dwarf, err
		}
		if upstream.Dispatch != "sent" || upstream.Code != "agent_pane_busy" && upstream.Code != "agent_name_taken" || !time.Now().Before(deadline) {
			return catalog.Character{}, mapHerdrError(err)
		}
		select {
		case <-time.After(agentStartRetryInterval):
		case <-ctx.Done():
			return catalog.Character{}, mapHerdrError(err)
		}
	}
}

// tryStartAgent makes one agent.start after re-reading that the pane still
// hosts the terminal. The name is the first dwarf, in the terminal's seeded
// order, whose name no live agent on this server holds; when every dwarf's
// name is held, the seeded dwarf's first free numbered variant (durinn-2).
func (manager *Manager) tryStartAgent(ctx context.Context, target TerminalTarget, kind string) (catalog.Character, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	pane, err := manager.resolvePane(ctx, target)
	if err != nil {
		return catalog.Character{}, err
	}
	agents, err := manager.agents(ctx)
	if err != nil {
		return catalog.Character{}, err
	}
	taken := make(map[string]bool, len(agents))
	for _, agent := range agents {
		taken[nameOf(agent)] = true
	}
	start := manager.seed(target.TerminalID)
	dwarf := manager.dwarves[start]
	name := agentName(dwarf)
	for offset := 1; taken[name] && offset < len(manager.dwarves); offset++ {
		dwarf = manager.dwarves[(start+offset)%len(manager.dwarves)]
		name = agentName(dwarf)
	}
	if taken[name] {
		dwarf = manager.dwarves[start]
		for number := 2; taken[name]; number++ {
			name = agentName(dwarf) + "-" + strconv.Itoa(number)
		}
	}
	var result struct {
		Type  string    `json:"type"`
		Agent agentInfo `json:"agent"`
	}
	if err := manager.herdr.Call(ctx, "agent.start", map[string]string{"name": name, "kind": kind, "pane_id": pane.ID}, &result); err != nil {
		return catalog.Character{}, err
	}
	if result.Type != "agent_started" || result.Agent.TerminalID != target.TerminalID || nameOf(result.Agent) != name {
		return catalog.Character{}, &Error{Code: ErrorOutcomeUnknown, Message: "The agent launch outcome is unknown.", Dispatch: "unknown"}
	}
	return dwarf, nil
}

// failedLaunch closes the terminal of a launch herdr definitely refused, so no
// half-made terminal is left behind, and reports the refusal. A launch whose
// outcome is unknown keeps its terminal, since an agent may be running there,
// as does one whose close does not succeed.
func (manager *Manager) failedLaunch(ctx context.Context, terminal Terminal, err error) error {
	var failure *Error
	if !errors.As(err, &failure) || failure.Dispatch == "unknown" {
		return createFailure(err, "identified", &terminal)
	}
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	pane, resolveErr := manager.resolvePane(ctx, terminal.Target)
	if resolveErr != nil {
		return createFailure(err, "identified", &terminal)
	}
	if _, closeErr := manager.closePane(ctx, pane.ID); closeErr != nil {
		return createFailure(err, "identified", &terminal)
	}
	copy := *failure
	copy.Dispatch = "sent"
	return &copy
}

func createFailure(err error, stage string, terminal *Terminal) error {
	var failure *Error
	if !errors.As(err, &failure) {
		failure = &Error{Code: ErrorOutcomeUnknown, Message: "Terminal creation may be incomplete.", Dispatch: "unknown"}
	}
	if stage != "" {
		copy := *failure
		if copy.Dispatch == "not_sent" {
			copy.Dispatch = "sent"
		}
		copy.Partial = &Partial{Stage: stage, Terminal: terminal}
		return &copy
	}
	return failure
}

func (manager *Manager) validateDestination(ctx context.Context, destination Destination, labelRequired bool) error {
	switch destination.Kind {
	case "existing":
		if destination.Label != "" {
			return newSessionError(ErrorNameInvalid, "Choose one exact workspace.")
		}
		return manager.resolveWorkspace(ctx, destination.WorkspaceTarget)
	case "new":
		if labelRequired && destination.Label == "" || destination.WorkspaceTarget != (WorkspaceTarget{}) {
			return newSessionError(ErrorNameInvalid, "Name the new workspace.")
		}
		if !validWorkspaceLabel(destination.Label) {
			return newSessionError(ErrorNameInvalid, "Use a valid workspace label.")
		}
		return nil
	default:
		return newSessionError(ErrorNameInvalid, "Choose an existing or new workspace.")
	}
}

func (manager *Manager) Rename(ctx context.Context, target TerminalTarget, name string) (ObservedTerminal, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	if err := validateName(name); err != nil {
		return ObservedTerminal{}, err
	}
	pane, err := manager.resolvePane(ctx, target)
	if err != nil {
		return ObservedTerminal{}, err
	}
	if err := manager.renamePane(ctx, pane.ID, name); err != nil {
		return ObservedTerminal{}, err
	}
	current, err := manager.resolvePane(ctx, target)
	if err != nil {
		return ObservedTerminal{}, afterMutation(err)
	}
	if current.Tokens["skid_named"] != "1" {
		if err := manager.reportPane(ctx, current.ID, map[string]string{"skid_named": "1"}); err != nil {
			failure := &Error{Code: ErrorMetadataUnavailable, Message: "The terminal was renamed but naming metadata was not retained.", Dispatch: "sent"}
			var upstream *Error
			if errors.As(err, &upstream) && upstream.Dispatch == "unknown" {
				failure = &Error{Code: ErrorOutcomeUnknown, Message: "The naming metadata outcome is unknown.", Dispatch: "unknown"}
			}
			if partial, observeErr := manager.observe(ctx, target); observeErr == nil {
				failure.Partial = &Partial{Terminal: &partial}
			}
			return ObservedTerminal{}, failure
		}
	}
	updated, err := manager.observe(ctx, target)
	if err != nil {
		return ObservedTerminal{}, afterMutation(err)
	}
	return ObservedTerminal{ObservedAt: time.Now().UTC(), Terminal: updated, Dispatch: "sent"}, nil
}

func (manager *Manager) renamePane(ctx context.Context, paneID, name string) error {
	var result struct {
		Type string   `json:"type"`
		Pane paneInfo `json:"pane"`
	}
	if err := manager.herdr.Call(ctx, "pane.rename", map[string]string{"pane_id": paneID, "label": name}, &result); err != nil {
		return mapHerdrError(err)
	}
	if result.Type != "pane_info" || result.Pane.ID != paneID || result.Pane.Label == nil || *result.Pane.Label != name {
		return &Error{Code: ErrorOutcomeUnknown, Message: "The rename outcome is unknown.", Dispatch: "unknown"}
	}
	return nil
}

func (manager *Manager) Move(ctx context.Context, target TerminalTarget, destination Destination) (ObservedTerminal, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	if err := manager.validateDestination(ctx, destination, true); err != nil {
		return ObservedTerminal{}, err
	}
	pane, err := manager.resolvePane(ctx, target)
	if err != nil {
		return ObservedTerminal{}, err
	}
	var moveDestination map[string]any
	if destination.Kind == "existing" {
		moveDestination = map[string]any{"type": "new_tab", "workspace_id": destination.WorkspaceTarget.WorkspaceID}
	} else {
		moveDestination = map[string]any{"type": "new_workspace", "label": destination.Label}
	}
	var result struct {
		Type string `json:"type"`
		Move struct {
			Pane paneInfo `json:"pane"`
		} `json:"move_result"`
	}
	if err := manager.herdr.Call(ctx, "pane.move", map[string]any{"pane_id": pane.ID, "destination": moveDestination, "focus": false}, &result); err != nil {
		return ObservedTerminal{}, mapHerdrError(err)
	}
	if result.Type != "pane_move" || result.Move.Pane.TerminalID != target.TerminalID {
		return ObservedTerminal{}, &Error{Code: ErrorOutcomeUnknown, Message: "The move outcome is unknown.", Dispatch: "unknown"}
	}
	updated, err := manager.observe(ctx, target)
	if err != nil {
		return ObservedTerminal{}, afterMutation(err)
	}
	return ObservedTerminal{ObservedAt: time.Now().UTC(), Terminal: updated, Dispatch: "sent"}, nil
}

func (manager *Manager) Kill(ctx context.Context, target TerminalTarget) (Closed, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	pane, err := manager.resolvePane(ctx, target)
	if err != nil {
		return Closed{}, err
	}
	return manager.closePane(ctx, pane.ID)
}

func (manager *Manager) closePane(ctx context.Context, paneID string) (Closed, error) {
	var result struct {
		Type string `json:"type"`
	}
	if err := manager.herdr.Call(ctx, "pane.close", map[string]string{"pane_id": paneID}, &result); err != nil {
		return Closed{}, mapHerdrError(err)
	}
	if result.Type != "ok" {
		return Closed{}, &Error{Code: ErrorOutcomeUnknown, Message: "The close outcome is unknown.", Dispatch: "unknown"}
	}
	return Closed{Terminal: "closed", Dispatch: "sent"}, nil
}

func mapWorkingDirectoryError(err error) error {
	code, classified := workdir.ErrorCodeOf(err)
	if !classified {
		return err
	}
	switch code {
	case workdir.Invalid:
		return newSessionError(ErrorWorkingDirectoryInvalid, "Use an absolute directory path or ~/… without terminal controls.")
	case workdir.Unavailable:
		return newSessionError(ErrorWorkingDirectoryUnavailable, "That directory is unavailable.")
	default:
		return err
	}
}

func afterMutation(err error) error {
	var failure *Error
	if !errors.As(err, &failure) {
		return &Error{Code: ErrorOutcomeUnknown, Message: "The terminal's current state is unavailable.", Dispatch: "sent"}
	}
	copy := *failure
	copy.Dispatch = "sent"
	return &copy
}
