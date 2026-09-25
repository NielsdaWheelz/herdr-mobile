package gateway

import (
	"errors"

	"github.com/NielsdaWheelz/skidbladnir/internal/machine"
	"github.com/NielsdaWheelz/skidbladnir/internal/reference"
	"github.com/NielsdaWheelz/skidbladnir/internal/sessions"
)

var errReferenceMachine = errors.New("resource reference names another machine")

func terminalReference(handle machine.Handle, target sessions.TerminalTarget) (string, error) {
	return reference.Encode(reference.Value{Kind: "terminal", Machine: handle.String(), TerminalID: target.TerminalID})
}

func workspaceReference(handle machine.Handle, target sessions.WorkspaceTarget) (string, error) {
	return reference.Encode(reference.Value{Kind: "workspace", Machine: handle.String(), WorkspaceID: target.WorkspaceID})
}

func agentReference(handle machine.Handle, target sessions.AgentTarget) (string, error) {
	return reference.Encode(reference.Value{Kind: "agent", Machine: handle.String(),
		TerminalID: target.Terminal.TerminalID, AgentName: target.Name})
}

func decodeReference(encoded, kind string, handle machine.Handle) (reference.Value, error) {
	value, err := reference.Decode(encoded)
	if err != nil || value.Kind != kind {
		return reference.Value{}, reference.ErrInvalid
	}
	if value.Machine != handle.String() {
		return reference.Value{}, errReferenceMachine
	}
	return value, nil
}

func parseTerminalReference(encoded string, handle machine.Handle) (sessions.TerminalTarget, error) {
	value, err := decodeReference(encoded, "terminal", handle)
	return sessions.TerminalTarget{TerminalID: value.TerminalID}, err
}

func parseWorkspaceReference(encoded string, handle machine.Handle) (sessions.WorkspaceTarget, error) {
	value, err := decodeReference(encoded, "workspace", handle)
	return sessions.WorkspaceTarget{WorkspaceID: value.WorkspaceID}, err
}

func parseAgentReference(encoded string, handle machine.Handle) (sessions.AgentTarget, error) {
	value, err := decodeReference(encoded, "agent", handle)
	return sessions.AgentTarget{Terminal: sessions.TerminalTarget{TerminalID: value.TerminalID}, Name: value.AgentName}, err
}
