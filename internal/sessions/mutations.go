package sessions

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/NielsdaWheelz/skidbladnir/internal/agentruntime"
	"github.com/NielsdaWheelz/skidbladnir/internal/workdir"
)

func (manager *Manager) Create(ctx context.Context, input CreateInput) (Created, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	return manager.create(ctx, input)
}

func (manager *Manager) Shell(ctx context.Context, source TerminalTarget) (Created, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	resolved, err := manager.resolveTerminal(ctx, source)
	if err != nil {
		return Created{}, err
	}
	panes, err := manager.panes(ctx)
	if err != nil {
		return Created{}, err
	}
	var cwd string
	for _, pane := range panes {
		if pane.ID == resolved.PaneID {
			if pane.ForegroundCWD != nil {
				cwd = *pane.ForegroundCWD
			} else if pane.CWD != nil {
				cwd = *pane.CWD
			}
			break
		}
	}
	if cwd == "" {
		return Created{}, newSessionError(ErrorWorkingDirectoryUnavailable, "The source directory is unavailable.")
	}
	return manager.create(ctx, CreateInput{Kind: LaunchTerminal, CWD: cwd,
		Destination: Destination{Kind: "existing", WorkspaceTarget: resolved.Terminal.WorkspaceTarget}})
}

func (manager *Manager) create(ctx context.Context, input CreateInput) (Created, error) {
	candidate, err := manager.workdir.ParseCandidate(input.CWD)
	if err != nil {
		return Created{}, mapWorkingDirectoryError(err)
	}
	cwd, err := manager.workdir.ValidateStart(candidate)
	if err != nil {
		return Created{}, mapWorkingDirectoryError(err)
	}
	var profile agentruntime.Profile
	namePrefix := "terminal"
	switch input.Kind {
	case LaunchAgent:
		var found bool
		profile, found = manager.profilesByKey[agentruntime.ProfileKey(input.Profile)]
		if !found {
			return Created{}, newSessionError(ErrorProfileUnknown, "Choose an available profile.")
		}
		namePrefix = string(profile.Key)
	case LaunchTerminal:
		if input.Profile != "" || input.Objective != "" {
			return Created{}, newSessionError(ErrorProfileUnknown, "A terminal has no agent profile or objective.")
		}
	default:
		return Created{}, newSessionError(ErrorProfileUnknown, "Choose a terminal or configured agent profile.")
	}
	name := input.Name
	if name == "" {
		name, err = manager.generatedName(ctx, namePrefix)
		if err != nil {
			return Created{}, err
		}
	}
	if err := validateName(name); err != nil {
		return Created{}, err
	}
	if err := validateObjective(input.Objective); err != nil {
		return Created{}, err
	}
	destination := input.Destination
	if destination.Kind == "" {
		destination.Kind = "new"
	}
	if destination.Kind == "new" && destination.Label == "" {
		destination.Label = name
	}
	if err := manager.validateDestination(ctx, destination, false); err != nil {
		return Created{}, err
	}
	if _, err := manager.workdir.ValidateStart(candidate); err != nil {
		return Created{}, mapWorkingDirectoryError(err)
	}
	env := make(map[string]string, len(profile.Environment))
	for _, variable := range profile.Environment {
		env[variable.Name] = variable.Value
	}
	var created struct {
		Type      string        `json:"type"`
		Workspace workspaceInfo `json:"workspace"`
		RootPane  paneInfo      `json:"root_pane"`
	}
	if destination.Kind == "existing" {
		var tab struct {
			Type     string   `json:"type"`
			RootPane paneInfo `json:"root_pane"`
		}
		err = manager.herdr.Call(ctx, "tab.create", map[string]any{"workspace_id": destination.WorkspaceTarget.WorkspaceID,
			"cwd": cwd.String(), "focus": false, "env": env}, &tab)
		if err == nil && (tab.Type != "tab_created" || tab.RootPane.ID == "" || tab.RootPane.TerminalID == "") {
			err = errors.New("invalid herdr tab creation response")
		}
		created.RootPane = tab.RootPane
	} else {
		err = manager.herdr.Call(ctx, "workspace.create", map[string]any{"cwd": cwd.String(), "focus": false,
			"label": destination.Label, "env": env}, &created)
		if err == nil && (created.Type != "workspace_created" || created.Workspace.ID == "" || created.RootPane.ID == "" || created.RootPane.TerminalID == "") {
			err = errors.New("invalid herdr workspace creation response")
		}
	}
	if err != nil {
		return Created{}, createFailure(err, "", nil)
	}
	pane := created.RootPane
	// A known create has happened. Every subsequent failure retains that fact.
	if err := manager.renamePane(ctx, pane.ID, name); err != nil {
		return Created{}, createFailure(err, "resource_created", nil)
	}
	metadata := map[string]string{"skid_named": "1"}
	if input.Kind == LaunchAgent {
		metadata["skid_launch_profile"] = string(profile.Key)
	}
	if err := manager.reportPane(ctx, pane.ID, metadata); err != nil {
		return Created{}, createFailure(err, "resource_created", nil)
	}
	if input.Objective != "" {
		if err := manager.writeObjective(ctx, pane.ID, input.Objective); err != nil {
			return Created{}, createFailure(err, "resource_created", nil)
		}
	}
	if destination.Kind == "new" {
		if _, err := manager.claimWorkspace(ctx, created.Workspace); err != nil {
			return Created{}, createFailure(err, "resource_created", nil)
		}
	}
	target, err := manager.claimPane(ctx, pane)
	if err != nil {
		return Created{}, createFailure(err, "resource_created", nil)
	}
	resolved, err := manager.resolveTerminal(ctx, target)
	if err != nil {
		return Created{}, createFailure(err, "resource_created", nil)
	}
	if resolved.Terminal.Name != name || input.Objective != "" && resolved.Terminal.Objective != input.Objective {
		return Created{}, createFailure(errors.New("terminal metadata was not retained"), "identified", &resolved.Terminal)
	}
	if input.Kind == LaunchAgent {
		command := "exec " + shellQuote(profile.Command)
		for _, argument := range agentruntime.LaunchArguments(profile, name) {
			command += " " + shellQuote(argument)
		}
		var result struct {
			Type string `json:"type"`
		}
		if err := manager.herdr.Call(ctx, "pane.send_input", map[string]any{"pane_id": resolved.PaneID,
			"text": command, "keys": []string{"enter"}}, &result); err != nil {
			return Created{}, createFailure(err, "identified", &resolved.Terminal)
		}
		if result.Type != "ok" {
			return Created{}, createFailure(errors.New("invalid launch response"), "identified", &resolved.Terminal)
		}
		return Created{ObservedAt: time.Now().UTC(), Terminal: resolved.Terminal, Launch: "submitted", Dispatch: "sent"}, nil
	}
	return Created{ObservedAt: time.Now().UTC(), Terminal: resolved.Terminal, Launch: "not_requested", Dispatch: "sent"}, nil
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
		_, err := manager.resolveWorkspace(ctx, destination.WorkspaceTarget)
		return err
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
	resolved, err := manager.resolveTerminal(ctx, target)
	if err != nil {
		return ObservedTerminal{}, err
	}
	if err := manager.renamePane(ctx, resolved.PaneID, name); err != nil {
		return ObservedTerminal{}, err
	}
	current, err := manager.pane(ctx, resolved.PaneID)
	if err != nil {
		return ObservedTerminal{}, afterMutation(err)
	}
	if current.Tokens["skid_named"] != "1" {
		if err := manager.reportPane(ctx, resolved.PaneID, map[string]string{"skid_named": "1"}); err != nil {
			partial := manager.project(ctx, current, target, resolved.Terminal.WorkspaceTarget)
			var failure *Error
			if errors.As(err, &failure) && failure.Dispatch == "unknown" {
				return ObservedTerminal{}, &Error{Code: ErrorOutcomeUnknown, Message: "The naming metadata outcome is unknown.", Dispatch: "unknown",
					Partial: &Partial{Terminal: &partial}}
			}
			return ObservedTerminal{}, &Error{Code: ErrorMetadataUnavailable, Message: "The terminal was renamed but naming metadata was not retained.", Dispatch: "sent",
				Partial: &Partial{Terminal: &partial}}
		}
	}
	updated, err := manager.resolveTerminal(ctx, target)
	if err != nil {
		return ObservedTerminal{}, afterMutation(err)
	}
	return ObservedTerminal{ObservedAt: time.Now().UTC(), Terminal: updated.Terminal, Dispatch: "sent"}, nil
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
	resolved, err := manager.resolveTerminal(ctx, target)
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
	if err := manager.herdr.Call(ctx, "pane.move", map[string]any{"pane_id": resolved.PaneID, "destination": moveDestination, "focus": false}, &result); err != nil {
		return ObservedTerminal{}, mapHerdrError(err)
	}
	if result.Type != "pane_move" || result.Move.Pane.TerminalID != target.TerminalID {
		return ObservedTerminal{}, &Error{Code: ErrorOutcomeUnknown, Message: "The move outcome is unknown.", Dispatch: "unknown"}
	}
	if destination.Kind == "new" {
		workspaces, err := manager.workspaces(ctx)
		if err != nil {
			return ObservedTerminal{}, afterMutation(err)
		}
		for _, workspace := range workspaces {
			if workspace.ID == result.Move.Pane.WorkspaceID {
				if _, err := manager.claimWorkspace(ctx, workspace); err != nil {
					return ObservedTerminal{}, &Error{Code: ErrorMetadataUnavailable, Message: "The moved terminal workspace is unaddressable.", Dispatch: "sent"}
				}
				break
			}
		}
	}
	updated, err := manager.resolveTerminal(ctx, target)
	if err != nil {
		return ObservedTerminal{}, afterMutation(err)
	}
	return ObservedTerminal{ObservedAt: time.Now().UTC(), Terminal: updated.Terminal, Dispatch: "sent"}, nil
}

func (manager *Manager) Kill(ctx context.Context, target TerminalTarget) (Closed, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	resolved, err := manager.resolveTerminal(ctx, target)
	if err != nil {
		return Closed{}, err
	}
	var result struct {
		Type string `json:"type"`
	}
	if err := manager.herdr.Call(ctx, "pane.close", map[string]string{"pane_id": resolved.PaneID}, &result); err != nil {
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

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
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
