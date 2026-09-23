package agentcontrol

import (
	"context"
	"errors"
	"unicode/utf8"

	"github.com/NielsdaWheelz/skidbladnir/internal/agentruntime"
	"github.com/NielsdaWheelz/skidbladnir/internal/herdr"
	"github.com/NielsdaWheelz/skidbladnir/internal/sessions"
)

var ErrInvalidInput = errors.New("invalid agent input")

type Service struct {
	sessions *sessions.Manager
	herdr    *herdr.Client
}

func New(manager *sessions.Manager, client *herdr.Client) (*Service, error) {
	if manager == nil || client == nil {
		return nil, errors.New("agent control requires sessions and herdr")
	}
	return &Service{sessions: manager, herdr: client}, nil
}

type ReadResult struct {
	Text      string `json:"text"`
	Source    string `json:"source"`
	Scope     string `json:"scope"`
	Truncated bool   `json:"truncated"`
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

// Enrich uses only Herdr's public agent state and explanation fields. It never
// reads previews or infers readiness from terminal text.
func (service *Service) Enrich(ctx context.Context, terminal *sessions.Terminal) {
	if terminal == nil || terminal.Agent == nil {
		return
	}
	agent := terminal.Agent
	resolved, err := service.sessions.ResolveAgent(ctx, agent.Target)
	if err != nil {
		agent.Status = agentruntime.Status{State: "unknown", Source: "unavailable", Reason: "observation_failed"}
		agent.Readiness = "unconfirmed"
		return
	}
	status, seq, ok := service.status(ctx, resolved.PaneID)
	if !ok {
		agent.Status = agentruntime.Status{State: "unknown", Source: "unavailable", Reason: "observation_failed"}
		agent.Readiness = "unconfirmed"
		return
	}
	agent.Status = status
	agent.Readiness = "unconfirmed"
	var explanation struct {
		Type    string `json:"type"`
		Explain struct {
			MatchedRule *struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"matched_rule"`
			VisibleIdle            bool    `json:"visible_idle"`
			VisibleBlocker         bool    `json:"visible_blocker"`
			ScreenDetectionSkipped bool    `json:"screen_detection_skipped"`
			FallbackReason         *string `json:"fallback_reason"`
		} `json:"explain"`
	}
	if err := service.herdr.Call(ctx, "agent.explain", map[string]string{"target": resolved.PaneID}, &explanation); err != nil || explanation.Type != "agent_explain" {
		return
	}
	if explanation.Explain.FallbackReason != nil && *explanation.Explain.FallbackReason == "default_known_agent_idle_fallback" {
		agent.Status.Reason = "default_idle"
	}
	if _, err := service.sessions.ResolveAgent(ctx, agent.Target); err != nil {
		agent.Status = agentruntime.Status{State: "unknown", Source: "unavailable", Reason: "observation_failed"}
		return
	}
	latest, latestSeq, ok := service.status(ctx, resolved.PaneID)
	if !ok || latest != status || latestSeq != seq {
		agent.Status = agentruntime.Status{State: "unknown", Source: "unavailable", Reason: "observation_failed"}
		return
	}
	if explanation.Explain.ScreenDetectionSkipped || explanation.Explain.FallbackReason != nil || explanation.Explain.MatchedRule == nil {
		return
	}
	if status.State == "idle" && explanation.Explain.VisibleIdle && explanation.Explain.MatchedRule.State == "idle" {
		agent.Readiness = "ready"
	} else if status.State == "blocked" && explanation.Explain.VisibleBlocker && explanation.Explain.MatchedRule.State == "blocked" {
		agent.Readiness = "blocked"
	}
}

func (service *Service) status(ctx context.Context, paneID string) (agentruntime.Status, uint64, bool) {
	var result struct {
		Type  string `json:"type"`
		Agent struct {
			PaneID string `json:"pane_id"`
			Status string `json:"agent_status"`
			Seq    uint64 `json:"state_change_seq"`
		} `json:"agent"`
	}
	if err := service.herdr.Call(ctx, "agent.get", map[string]string{"target": paneID}, &result); err != nil || result.Type != "agent_info" || result.Agent.PaneID != paneID {
		return agentruntime.Status{}, 0, false
	}
	status := agentruntime.Status{Source: "herdr"}
	switch result.Agent.Status {
	case "working", "blocked", "idle":
		status.State = result.Agent.Status
	case "done":
		status.State = "idle"
	case "unknown":
		status.State = "unknown"
		status.Reason = "unrecognized"
	default:
		status.State = "unknown"
		status.Reason = "unrecognized"
	}
	return status, result.Agent.Seq, true
}

func (service *Service) Read(ctx context.Context, target sessions.AgentTarget, coverage string, maxBytes int) (ReadResult, error) {
	if coverage == "" {
		coverage = "recent"
	}
	if maxBytes == 0 {
		maxBytes = 16384
	}
	if coverage != "recent" && coverage != "visible" || maxBytes < 1 || maxBytes > 32768 {
		return ReadResult{}, ErrInvalidInput
	}
	resolved, err := service.sessions.ResolveAgent(ctx, target)
	if err != nil {
		return ReadResult{}, err
	}
	source, scope := "recent_unwrapped", "terminal_history"
	if coverage == "visible" {
		source, scope = "visible", "visible"
	}
	var result struct {
		Type string `json:"type"`
		Read struct {
			PaneID    string `json:"pane_id"`
			Source    string `json:"source"`
			Text      string `json:"text"`
			Truncated bool   `json:"truncated"`
		} `json:"read"`
	}
	if err := service.herdr.Call(ctx, "pane.read", map[string]any{"pane_id": resolved.PaneID, "source": source,
		"lines": 1000, "format": "text", "strip_ansi": true}, &result); err != nil {
		return ReadResult{}, controlFailure(err)
	}
	if result.Type != "pane_read" || result.Read.PaneID != resolved.PaneID || result.Read.Source != source || !utf8.ValidString(result.Read.Text) {
		return ReadResult{}, &sessions.Error{Code: sessions.ErrorOutcomeUnknown, Message: "The terminal read was invalid.", Dispatch: "unknown"}
	}
	text := result.Read.Text
	truncated := result.Read.Truncated
	if len(text) > maxBytes {
		start := len(text) - maxBytes
		for start < len(text) && !utf8.RuneStart(text[start]) {
			start++
		}
		text, truncated = text[start:], true
	}
	return ReadResult{Text: text, Source: "terminal", Scope: scope, Truncated: truncated}, nil
}
