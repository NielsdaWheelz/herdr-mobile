// Package reference owns the opaque wire format for exact resource targets.
// A reference is a thin encoding of herdr's own identities on one machine:
// a terminal's terminal_id, a workspace's id, and an agent's terminal_id plus
// its herdr agent name. herdr never repeats a terminal_id, so a terminal or
// agent reference cannot name a later terminal; an agent herdr did not name
// has no reference.
package reference

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/NielsdaWheelz/skidbladnir/internal/machine"
	"github.com/NielsdaWheelz/skidbladnir/internal/strictjson"
)

type Value struct {
	Kind        string `json:"kind"`
	Machine     string `json:"machine"`
	TerminalID  string `json:"terminalId,omitempty"`
	WorkspaceID string `json:"workspaceId,omitempty"`
	AgentName   string `json:"agentName,omitempty"`
}

var ErrInvalid = errors.New("invalid resource reference")

const maximumLength = 4096

func Encode(value Value) (string, error) {
	if !value.valid() {
		return "", ErrInvalid
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", ErrInvalid
	}
	encoded := base64.RawURLEncoding.EncodeToString(data)
	if len(encoded) > maximumLength {
		return "", ErrInvalid
	}
	return encoded, nil
}

// Decode accepts only the exact canonical encoding of a valid value, so a
// reference has one spelling.
func Decode(encoded string) (Value, error) {
	if encoded == "" || len(encoded) > maximumLength {
		return Value{}, ErrInvalid
	}
	data, err := base64.RawURLEncoding.Strict().DecodeString(encoded)
	if err != nil {
		return Value{}, ErrInvalid
	}
	var value Value
	if strictjson.Decode(data, &value) != nil {
		return Value{}, ErrInvalid
	}
	if canonical, err := Encode(value); err != nil || canonical != encoded {
		return Value{}, ErrInvalid
	}
	return value, nil
}

func (value Value) valid() bool {
	if _, err := machine.Parse(value.Machine); err != nil {
		return false
	}
	switch value.Kind {
	case "terminal":
		return value.TerminalID != "" && value.WorkspaceID == "" && value.AgentName == ""
	case "workspace":
		return value.WorkspaceID != "" && value.TerminalID == "" && value.AgentName == ""
	case "agent":
		return value.TerminalID != "" && value.AgentName != "" && value.WorkspaceID == ""
	default:
		return false
	}
}
