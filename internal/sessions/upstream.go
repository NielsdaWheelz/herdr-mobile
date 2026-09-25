package sessions

import (
	"context"
	"errors"

	"github.com/NielsdaWheelz/skidbladnir/internal/herdr"
)

const metadataSource = "user:skidbladnir"

type workspaceInfo struct {
	ID    string `json:"workspace_id"`
	Label string `json:"label"`
}

type paneInfo struct {
	ID            string            `json:"pane_id"`
	TerminalID    string            `json:"terminal_id"`
	WorkspaceID   string            `json:"workspace_id"`
	CWD           *string           `json:"cwd"`
	ForegroundCWD *string           `json:"foreground_cwd"`
	Label         *string           `json:"label"`
	Agent         *string           `json:"agent"`
	Tokens        map[string]string `json:"tokens"`
}

type agentInfo struct {
	TerminalID string  `json:"terminal_id"`
	PaneID     string  `json:"pane_id"`
	Name       *string `json:"name"`
	Agent      *string `json:"agent"`
	Status     string  `json:"agent_status"`
	Seq        uint64  `json:"state_change_seq"`
}

func (manager *Manager) workspaces(ctx context.Context) ([]workspaceInfo, error) {
	var result struct {
		Type       string          `json:"type"`
		Workspaces []workspaceInfo `json:"workspaces"`
	}
	if err := manager.herdr.Call(ctx, "workspace.list", struct{}{}, &result); err != nil {
		return nil, mapReadHerdrError(err)
	}
	if result.Type != "workspace_list" || result.Workspaces == nil {
		return nil, errors.New("invalid herdr workspace inventory")
	}
	seen := make(map[string]bool, len(result.Workspaces))
	for _, workspace := range result.Workspaces {
		if workspace.ID == "" || seen[workspace.ID] {
			return nil, errors.New("invalid herdr workspace identity")
		}
		seen[workspace.ID] = true
	}
	return result.Workspaces, nil
}

func (manager *Manager) panes(ctx context.Context) ([]paneInfo, error) {
	var result struct {
		Type  string     `json:"type"`
		Panes []paneInfo `json:"panes"`
	}
	if err := manager.herdr.Call(ctx, "pane.list", struct{}{}, &result); err != nil {
		return nil, mapReadHerdrError(err)
	}
	if result.Type != "pane_list" || result.Panes == nil {
		return nil, errors.New("invalid herdr pane inventory")
	}
	seen := make(map[string]bool, len(result.Panes))
	terminals := make(map[string]bool, len(result.Panes))
	for _, pane := range result.Panes {
		if pane.ID == "" || pane.TerminalID == "" || pane.WorkspaceID == "" || seen[pane.ID] || terminals[pane.TerminalID] {
			return nil, errors.New("invalid herdr pane identity")
		}
		seen[pane.ID], terminals[pane.TerminalID] = true, true
	}
	return result.Panes, nil
}

func (manager *Manager) pane(ctx context.Context, paneID string) (paneInfo, error) {
	var result struct {
		Type string   `json:"type"`
		Pane paneInfo `json:"pane"`
	}
	if err := manager.herdr.Call(ctx, "pane.get", map[string]string{"pane_id": paneID}, &result); err != nil {
		return paneInfo{}, mapReadHerdrError(err)
	}
	if result.Type != "pane_info" || result.Pane.ID != paneID || result.Pane.TerminalID == "" || result.Pane.WorkspaceID == "" {
		return paneInfo{}, errors.New("invalid herdr pane response")
	}
	return result.Pane, nil
}

// agent reads herdr's agent in one pane. found is false when herdr reports no
// agent there.
func (manager *Manager) agent(ctx context.Context, paneID string) (agentInfo, bool, error) {
	var result struct {
		Type  string    `json:"type"`
		Agent agentInfo `json:"agent"`
	}
	if err := manager.herdr.Call(ctx, "agent.get", map[string]string{"target": paneID}, &result); err != nil {
		var upstream *herdr.Error
		if errors.As(err, &upstream) && upstream.Dispatch == "sent" && upstream.Code == "agent_not_found" {
			return agentInfo{}, false, nil
		}
		return agentInfo{}, false, mapReadHerdrError(err)
	}
	if result.Type != "agent_info" || result.Agent.PaneID != paneID || result.Agent.TerminalID == "" {
		return agentInfo{}, false, errors.New("invalid herdr agent response")
	}
	return result.Agent, true, nil
}

func (manager *Manager) agents(ctx context.Context) ([]agentInfo, error) {
	var result struct {
		Type   string      `json:"type"`
		Agents []agentInfo `json:"agents"`
	}
	if err := manager.herdr.Call(ctx, "agent.list", struct{}{}, &result); err != nil {
		return nil, mapReadHerdrError(err)
	}
	if result.Type != "agent_list" || result.Agents == nil {
		return nil, errors.New("invalid herdr agent inventory")
	}
	return result.Agents, nil
}

func (manager *Manager) reportPane(ctx context.Context, paneID string, tokens map[string]string) error {
	var result struct {
		Type string `json:"type"`
	}
	if err := manager.herdr.Call(ctx, "pane.report_metadata", map[string]any{"pane_id": paneID, "source": metadataSource, "tokens": tokens}, &result); err != nil {
		return mapHerdrError(err)
	}
	if result.Type != "ok" {
		return errors.New("invalid herdr metadata response")
	}
	return nil
}

func mapHerdrError(err error) *Error {
	var upstream *herdr.Error
	if !errors.As(err, &upstream) {
		return &Error{Code: ErrorHerdrUnavailable, Message: "Herdr is unavailable.", Dispatch: "unknown"}
	}
	if upstream.Dispatch == "not_sent" {
		return &Error{Code: ErrorHerdrUnavailable, Message: "Herdr is unavailable.", Dispatch: "not_sent"}
	}
	if upstream.Dispatch == "unknown" {
		return &Error{Code: ErrorOutcomeUnknown, Message: "The upstream outcome is unknown.", Dispatch: "unknown"}
	}
	switch upstream.Code {
	case "confirmation_required":
		return &Error{Code: ErrorClosureConfirmationRequired, Message: "Herdr refused terminal closure.", Dispatch: upstream.Dispatch, Partial: &Partial{TerminalOutcome: "refused"}}
	case "pane_not_found":
		return &Error{Code: ErrorTerminalStale, Message: "The selected terminal changed.", Dispatch: upstream.Dispatch}
	case "workspace_not_found":
		return &Error{Code: ErrorWorkspaceStale, Message: "The selected workspace changed.", Dispatch: upstream.Dispatch}
	case "agent_not_found", "agent_not_ready":
		return &Error{Code: ErrorAgentStale, Message: "The selected agent changed.", Dispatch: upstream.Dispatch}
	default:
		return &Error{Code: ErrorUpstreamRejected, Message: "Herdr rejected the operation.", Dispatch: upstream.Dispatch}
	}
}

// A failed observation cannot establish whether the read completed, but no
// native mutation was attempted by that read.
func mapReadHerdrError(err error) error {
	failure := mapHerdrError(err)
	failure.Dispatch = "not_sent"
	return failure
}
