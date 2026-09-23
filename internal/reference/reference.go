// Package reference owns the opaque wire format for exact resource targets.
package reference

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/NielsdaWheelz/skidbladnir/internal/machine"
	"github.com/NielsdaWheelz/skidbladnir/internal/strictjson"
)

type Agent struct {
	PID                int    `json:"pid"`
	StartIdentity      string `json:"startIdentity"`
	CommandFingerprint string `json:"commandFingerprint"`
	Provider           string `json:"provider"`
}

type Value struct {
	Kind          string `json:"kind"`
	Machine       string `json:"machine"`
	TerminalID    string `json:"terminalId,omitempty"`
	WorkspaceID   string `json:"workspaceId,omitempty"`
	IdentityToken string `json:"identityToken"`
	Agent         *Agent `json:"agent,omitempty"`
}

var ErrInvalid = errors.New("invalid resource reference")

func Encode(value Value) (string, error) {
	if !value.valid() {
		return "", ErrInvalid
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", ErrInvalid
	}
	encoded := base64.RawURLEncoding.EncodeToString(data)
	if len(encoded) > 4096 {
		return "", ErrInvalid
	}
	return encoded, nil
}

func Decode(encoded string) (Value, error) {
	if encoded == "" || len(encoded) > 4096 {
		return Value{}, ErrInvalid
	}
	data, err := base64.RawURLEncoding.Strict().DecodeString(encoded)
	if err != nil || base64.RawURLEncoding.EncodeToString(data) != encoded {
		return Value{}, ErrInvalid
	}
	var value Value
	if strictjson.Decode(data, &value) != nil || !value.valid() {
		return Value{}, ErrInvalid
	}
	var fields map[string]json.RawMessage
	if strictjson.Decode(data, &fields) != nil {
		return Value{}, ErrInvalid
	}
	for _, field := range fields {
		if string(field) == "null" {
			return Value{}, ErrInvalid
		}
	}
	expected := map[string]bool{"kind": true, "machine": true, "identityToken": true}
	switch value.Kind {
	case "terminal":
		expected["terminalId"] = true
	case "workspace":
		expected["workspaceId"] = true
	case "agent":
		expected["terminalId"] = true
		expected["agent"] = true
	}
	if len(fields) != len(expected) {
		return Value{}, ErrInvalid
	}
	for name := range fields {
		if !expected[name] {
			return Value{}, ErrInvalid
		}
	}
	if value.Agent != nil {
		var agentFields map[string]json.RawMessage
		if strictjson.Decode(fields["agent"], &agentFields) != nil {
			return Value{}, ErrInvalid
		}
		if len(agentFields) != 4 {
			return Value{}, ErrInvalid
		}
		for _, field := range agentFields {
			if string(field) == "null" {
				return Value{}, ErrInvalid
			}
		}
	}
	return value, nil
}

func (value Value) valid() bool {
	if _, err := machine.Parse(value.Machine); err != nil {
		return false
	}
	token, err := base64.RawURLEncoding.Strict().DecodeString(value.IdentityToken)
	if err != nil || len(token) != 16 || base64.RawURLEncoding.EncodeToString(token) != value.IdentityToken {
		return false
	}
	switch value.Kind {
	case "terminal":
		return value.TerminalID != "" && value.WorkspaceID == "" && value.Agent == nil
	case "workspace":
		return value.WorkspaceID != "" && value.TerminalID == "" && value.Agent == nil
	case "agent":
		if value.TerminalID == "" || value.WorkspaceID != "" || value.Agent == nil {
			return false
		}
		agent := value.Agent
		if agent.PID <= 0 || agent.StartIdentity == "" || (agent.Provider != "Codex" && agent.Provider != "Claude") || len(agent.CommandFingerprint) != 64 {
			return false
		}
		return strings.Trim(agent.CommandFingerprint, "0123456789abcdef") == ""
	default:
		return false
	}
}
