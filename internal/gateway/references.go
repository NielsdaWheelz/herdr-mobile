package gateway

import (
	"errors"

	"github.com/NielsdaWheelz/skidbladnir/internal/agentruntime"
	"github.com/NielsdaWheelz/skidbladnir/internal/machine"
	"github.com/NielsdaWheelz/skidbladnir/internal/process"
	"github.com/NielsdaWheelz/skidbladnir/internal/reference"
	"github.com/NielsdaWheelz/skidbladnir/internal/sessions"
)

var errReferenceMachine = errors.New("resource reference names another machine")

func terminalReference(handle machine.Handle, target sessions.TerminalTarget) (string, error) {
	return reference.Encode(reference.Value{
		Kind: "terminal", Machine: handle.String(), TerminalID: target.TerminalID, IdentityToken: target.IdentityToken,
	})
}

func workspaceReference(handle machine.Handle, target sessions.WorkspaceTarget) (string, error) {
	return reference.Encode(reference.Value{
		Kind: "workspace", Machine: handle.String(), WorkspaceID: target.WorkspaceID, IdentityToken: target.IdentityToken,
	})
}

func agentReference(handle machine.Handle, target sessions.AgentTarget) (string, error) {
	return reference.Encode(reference.Value{
		Kind: "agent", Machine: handle.String(), TerminalID: target.TerminalTarget.TerminalID,
		IdentityToken: target.TerminalTarget.IdentityToken,
		Agent: &reference.Agent{PID: int(target.PID), StartIdentity: string(target.StartIdentity),
			CommandFingerprint: target.CommandFingerprint, Provider: string(target.Provider)},
	})
}

func parseTerminalReference(encoded string, handle machine.Handle) (sessions.TerminalTarget, error) {
	value, err := reference.Decode(encoded)
	if err != nil || value.Kind != "terminal" {
		return sessions.TerminalTarget{}, reference.ErrInvalid
	}
	if value.Machine != handle.String() {
		return sessions.TerminalTarget{}, errReferenceMachine
	}
	return sessions.TerminalTarget{TerminalID: value.TerminalID, IdentityToken: value.IdentityToken}, nil
}

func parseWorkspaceReference(encoded string, handle machine.Handle) (sessions.WorkspaceTarget, error) {
	value, err := reference.Decode(encoded)
	if err != nil || value.Kind != "workspace" {
		return sessions.WorkspaceTarget{}, reference.ErrInvalid
	}
	if value.Machine != handle.String() {
		return sessions.WorkspaceTarget{}, errReferenceMachine
	}
	return sessions.WorkspaceTarget{WorkspaceID: value.WorkspaceID, IdentityToken: value.IdentityToken}, nil
}

func parseAgentReference(encoded string, handle machine.Handle) (sessions.AgentTarget, error) {
	value, err := reference.Decode(encoded)
	if err != nil || value.Kind != "agent" {
		return sessions.AgentTarget{}, reference.ErrInvalid
	}
	if value.Machine != handle.String() {
		return sessions.AgentTarget{}, errReferenceMachine
	}
	return sessions.AgentTarget{
		TerminalTarget:     sessions.TerminalTarget{TerminalID: value.TerminalID, IdentityToken: value.IdentityToken},
		PID:                process.PID(value.Agent.PID),
		StartIdentity:      process.StartIdentity(value.Agent.StartIdentity),
		CommandFingerprint: value.Agent.CommandFingerprint,
		Provider:           agentruntime.Provider(value.Agent.Provider),
	}, nil
}
