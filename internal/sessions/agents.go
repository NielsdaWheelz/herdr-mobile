package sessions

import (
	"context"
	"errors"
)

func (manager *Manager) Interrupt(ctx context.Context, target AgentTarget) (WriteResult, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	return manager.interrupt(ctx, target)
}

// Stop interrupts once, re-checks the original target, then closes its
// terminal. The same agent, or no agent because it has exited, permits the
// close; another agent in that terminal does not. The re-check and the close
// are not atomic against another herdr client.
func (manager *Manager) Stop(ctx context.Context, target AgentTarget) (StopResult, error) {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	if _, err := manager.interrupt(ctx, target); err != nil {
		var failure *Error
		if errors.As(err, &failure) && failure.Dispatch != "not_sent" {
			copy := *failure
			copy.Partial = &Partial{AgentOutcome: "unconfirmed", TerminalOutcome: "not_attempted"}
			return StopResult{}, &copy
		}
		return StopResult{}, err
	}
	agentOutcome := "interrupt_sent"
	pane, err := manager.resolvePane(ctx, target.Terminal)
	if err != nil {
		return StopResult{}, stopFailure(err, agentOutcome, false)
	}
	agent, found, err := manager.agent(ctx, pane.ID)
	if err != nil {
		return StopResult{}, stopFailure(err, agentOutcome, false)
	}
	if !found {
		agentOutcome = "exited"
	} else if !sameAgent(agent, target) {
		return StopResult{}, stopFailure(&Error{Code: ErrorAgentStale, Message: "The selected agent changed.", Dispatch: "not_sent"}, agentOutcome, false)
	}
	closed, err := manager.closePane(ctx, pane.ID)
	if err != nil {
		return StopResult{}, stopFailure(err, agentOutcome, true)
	}
	return StopResult{Agent: agentOutcome, Terminal: closed.Terminal, Dispatch: closed.Dispatch}, nil
}

// interrupt re-reads the target's terminal and requires that herdr still
// reports the same named agent there, then sends that provider's interrupt key
// to the agent by its herdr name, so herdr binds the write to that agent and
// refuses it if the agent is no longer its terminal's foreground process.
func (manager *Manager) interrupt(ctx context.Context, target AgentTarget) (WriteResult, error) {
	pane, err := manager.resolvePane(ctx, target.Terminal)
	if err != nil {
		return WriteResult{}, err
	}
	agent, found, err := manager.agent(ctx, pane.ID)
	if err != nil {
		return WriteResult{}, err
	}
	if !found || !sameAgent(agent, target) || agent.Agent == nil {
		return WriteResult{}, &Error{Code: ErrorAgentStale, Message: "The selected agent changed.", Dispatch: "not_sent"}
	}
	key := "ctrl+c"
	if *agent.Agent == "codex" {
		key = "escape"
	}
	var result struct {
		Type string `json:"type"`
	}
	if err := manager.herdr.Call(ctx, "agent.send_keys", map[string]any{"target": target.Name, "keys": []string{key}}, &result); err != nil {
		return WriteResult{}, mapHerdrError(err)
	}
	if result.Type != "ok" {
		return WriteResult{}, &Error{Code: ErrorOutcomeUnknown, Message: "Input delivery is unknown.", Dispatch: "unknown"}
	}
	return WriteResult{Outcome: "written", Dispatch: "sent"}, nil
}

// stopFailure reports a failure after an acknowledged interrupt, which makes
// the whole stop at least sent. The close error's own dispatch classifies the
// close: none dispatched is not_attempted, an acknowledged rejection is
// refused, and a lost reply is unconfirmed.
func stopFailure(err error, agentOutcome string, closeAttempted bool) error {
	var failure *Error
	if !errors.As(err, &failure) {
		return &Error{Code: ErrorOutcomeUnknown, Message: "The terminal's current state is unavailable.", Dispatch: "sent",
			Partial: &Partial{AgentOutcome: agentOutcome, TerminalOutcome: "not_attempted"}}
	}
	copy := *failure
	terminalOutcome := "not_attempted"
	if closeAttempted {
		switch failure.Dispatch {
		case "not_sent":
		case "sent":
			terminalOutcome = "refused"
		case "unknown":
			terminalOutcome = "unconfirmed"
		default:
			panic("invalid close dispatch") // justify-defect: sessions errors carry only not_sent, sent or unknown dispatch.
		}
	}
	if copy.Dispatch == "not_sent" {
		copy.Dispatch = "sent"
	}
	copy.Partial = &Partial{AgentOutcome: agentOutcome, TerminalOutcome: terminalOutcome}
	return &copy
}
