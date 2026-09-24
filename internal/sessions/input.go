package sessions

import (
	"context"
	"strings"
)

// SendText types text into target's terminal.
func (manager *Manager) SendText(ctx context.Context, target TerminalTarget, text string) error {
	return manager.input(ctx, target, "pane.send_text", map[string]any{"text": text})
}

// Paste pastes text into target's terminal.
func (manager *Manager) Paste(ctx context.Context, target TerminalTarget, text string) error {
	return manager.input(ctx, target, "pane.send_input", map[string]any{"text": text})
}

// SendKey presses one key with its modifiers in target's terminal, spelled as
// herdr's key parser names it.
func (manager *Manager) SendKey(ctx context.Context, target TerminalTarget, key string, modifiers []string) error {
	switch key {
	case " ":
		key = "space"
	case "+":
		key = "plus"
	}
	if len(modifiers) > 0 {
		key = strings.Join(modifiers, "+") + "+" + key
	}
	return manager.input(ctx, target, "pane.send_input", map[string]any{"keys": []string{key}})
}

// input re-reads which pane hosts target's terminal and writes there.
func (manager *Manager) input(ctx context.Context, target TerminalTarget, method string, params map[string]any) error {
	manager.mutations.Lock()
	defer manager.mutations.Unlock()
	pane, err := manager.resolvePane(ctx, target)
	if err != nil {
		return err
	}
	params["pane_id"] = pane.ID
	var result struct {
		Type string `json:"type"`
	}
	if err := manager.herdr.Call(ctx, method, params, &result); err != nil {
		return mapHerdrError(err)
	}
	if result.Type != "ok" {
		return &Error{Code: ErrorOutcomeUnknown, Message: "Input delivery is unknown.", Dispatch: "unknown"}
	}
	return nil
}
