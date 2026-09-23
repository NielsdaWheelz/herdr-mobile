package agentcontrol

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/NielsdaWheelz/skidbladnir/internal/agentruntime"
	"github.com/NielsdaWheelz/skidbladnir/internal/herdr"
	"github.com/NielsdaWheelz/skidbladnir/internal/process"
	"github.com/NielsdaWheelz/skidbladnir/internal/sessions"
)

func (service *Service) Send(ctx context.Context, target sessions.AgentTarget, text, mode string) (WriteResult, error) {
	if text == "" || len(text) > 32768 || !utf8.ValidString(text) || strings.ContainsRune(text, 0) ||
		mode != "" && mode != "auto" && mode != "terminal" {
		return WriteResult{}, ErrInvalidInput
	}
	resolved, err := service.sessions.ResolveAgent(ctx, target)
	if err != nil {
		return WriteResult{}, err
	}
	if mode != "terminal" {
		terminal := resolved.Terminal
		service.Enrich(ctx, &terminal)
		if terminal.Agent == nil || terminal.Agent.Readiness != "ready" {
			return WriteResult{}, &sessions.Error{Code: sessions.ErrorReadinessUnconfirmed, Message: "Agent readiness is unconfirmed; use explicit terminal input or wait.", Dispatch: "not_sent"}
		}
	}
	resolved, err = service.sessions.ResolveAgent(ctx, target)
	if err != nil {
		return WriteResult{}, err
	}
	return service.write(ctx, resolved.PaneID, text, []string{"enter"})
}

func (service *Service) Keys(ctx context.Context, target sessions.AgentTarget, keys []string) (WriteResult, error) {
	if len(keys) < 1 || len(keys) > 16 {
		return WriteResult{}, ErrInvalidInput
	}
	for _, key := range keys {
		switch key {
		case "enter", "escape", "ctrl-c", "up", "down", "left", "right", "tab", "backspace":
		default:
			return WriteResult{}, ErrInvalidInput
		}
	}
	resolved, err := service.sessions.ResolveAgent(ctx, target)
	if err != nil {
		return WriteResult{}, err
	}
	return service.write(ctx, resolved.PaneID, "", keys)
}

func (service *Service) Interrupt(ctx context.Context, target sessions.AgentTarget) (WriteResult, error) {
	resolved, err := service.sessions.ResolveAgent(ctx, target)
	if err != nil {
		return WriteResult{}, err
	}
	key := "ctrl-c"
	if target.Provider == agentruntime.ProviderCodex {
		key = "escape"
	}
	return service.write(ctx, resolved.PaneID, "", []string{key})
}

func (service *Service) write(ctx context.Context, paneID, text string, keys []string) (WriteResult, error) {
	var result struct {
		Type string `json:"type"`
	}
	if err := service.herdr.Call(ctx, "pane.send_input", map[string]any{"pane_id": paneID, "text": text, "keys": keys}, &result); err != nil {
		return WriteResult{}, controlFailure(err)
	}
	if result.Type != "ok" {
		return WriteResult{}, &sessions.Error{Code: sessions.ErrorOutcomeUnknown, Message: "Input delivery is unknown.", Dispatch: "unknown"}
	}
	return WriteResult{Method: "terminal", Outcome: "written", Dispatch: "sent"}, nil
}

func (service *Service) Stop(ctx context.Context, target sessions.AgentTarget) (StopResult, error) {
	write, err := service.Interrupt(ctx, target)
	if err != nil {
		var failure *sessions.Error
		if errors.As(err, &failure) && failure.Dispatch != "not_sent" {
			copy := *failure
			copy.Partial = &sessions.Partial{AgentOutcome: "unconfirmed", TerminalOutcome: "not_attempted"}
			return StopResult{}, &copy
		}
		return StopResult{}, err
	}
	if write.Dispatch != "sent" {
		return StopResult{}, &sessions.Error{Code: sessions.ErrorOutcomeUnknown, Message: "Interrupt delivery is unknown.", Dispatch: "unknown",
			Partial: &sessions.Partial{AgentOutcome: "unconfirmed", TerminalOutcome: "not_attempted"}}
	}
	agentOutcome := "interrupt_sent"
	terminal, err := service.sessions.ResolveTerminal(ctx, target.TerminalTarget)
	if err != nil {
		return StopResult{}, stopFailure(err, agentOutcome, false)
	}
	if terminal.Terminal.Agent != nil && terminal.Terminal.Agent.Target != target {
		return StopResult{}, stopFailure(&sessions.Error{Code: sessions.ErrorAgentStale, Message: "The selected agent changed.", Dispatch: "not_sent"}, agentOutcome, false)
	}
	observed, observationErr := process.Observe(target.PID)
	if errors.Is(observationErr, process.ErrProcessAbsent) || observationErr == nil && observed.StartIdentity != target.StartIdentity {
		agentOutcome = "exited"
	} else if observationErr != nil {
		return StopResult{}, stopFailure(&sessions.Error{Code: sessions.ErrorAgentStale, Message: "The selected agent could not be revalidated.", Dispatch: "not_sent"}, agentOutcome, false)
	} else if terminal.Terminal.Agent == nil {
		return StopResult{}, stopFailure(&sessions.Error{Code: sessions.ErrorAgentStale, Message: "The selected agent is no longer foreground.", Dispatch: "not_sent"}, agentOutcome, false)
	}
	// The original terminal ref remains the closure target. Sessions owns the
	// final native close and its refusal policy; no broader primitive follows.
	closed, err := service.sessions.Kill(ctx, target.TerminalTarget)
	if err != nil {
		return StopResult{}, stopFailure(err, agentOutcome, true)
	}
	return StopResult{Agent: agentOutcome, Terminal: closed.Terminal, Dispatch: closed.Dispatch}, nil
}

func stopFailure(err error, agentOutcome string, closeAttempted bool) error {
	var failure *sessions.Error
	if !errors.As(err, &failure) {
		return err
	}
	copy := *failure
	if copy.Dispatch == "not_sent" {
		copy.Dispatch = "sent"
	}
	terminalOutcome := "not_attempted"
	if closeAttempted {
		if copy.Code == sessions.ErrorClosureConfirmationRequired {
			terminalOutcome = "refused"
		} else if copy.Dispatch == "unknown" {
			terminalOutcome = "unconfirmed"
		}
	}
	copy.Partial = &sessions.Partial{AgentOutcome: agentOutcome, TerminalOutcome: terminalOutcome}
	return &copy
}

func controlFailure(err error) error {
	var upstream *herdr.Error
	if !errors.As(err, &upstream) {
		return &sessions.Error{Code: sessions.ErrorHerdrUnavailable, Message: "Herdr is unavailable.", Dispatch: "not_sent"}
	}
	if upstream.Dispatch == "unknown" {
		return &sessions.Error{Code: sessions.ErrorOutcomeUnknown, Message: "The upstream outcome is unknown.", Dispatch: "unknown"}
	}
	if upstream.Code == "invalid_key" {
		return ErrInvalidInput
	}
	if upstream.Dispatch == "not_sent" {
		return &sessions.Error{Code: sessions.ErrorHerdrUnavailable, Message: "Herdr is unavailable.", Dispatch: "not_sent"}
	}
	return &sessions.Error{Code: sessions.ErrorUpstreamRejected, Message: "Herdr rejected the operation.", Dispatch: "sent"}
}
