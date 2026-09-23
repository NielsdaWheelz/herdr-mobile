package fleetclient

import (
	"unicode/utf8"

	"github.com/NielsdaWheelz/skidbladnir/internal/reference"
)

type LaunchKind string

const (
	LaunchAgent    LaunchKind = "agent"
	LaunchTerminal LaunchKind = "terminal"
)

// Request is the cli's parsed intent, never an HTTP body.
type Request struct {
	Operation, Name, Machine, Ref, NewName      string
	Kind                                        LaunchKind
	Profile, CWD, Objective                     string
	DestinationKind, WorkspaceRef, NewWorkspace string
	Text                                        string
	Keys                                        []string
	Mode, Coverage                              string
	MaxBytes                                    int
}

func (request Request) Valid() bool {
	if request.Operation == "" {
		return false
	}
	if request.Operation != "start" && (request.Kind != "" || request.Profile != "" || request.CWD != "" || request.Objective != "") {
		return false
	}
	if request.Operation != "start" && request.Operation != "move" && (request.DestinationKind != "" || request.WorkspaceRef != "" || request.NewWorkspace != "") {
		return false
	}
	if request.Operation != "rename" && request.NewName != "" {
		return false
	}
	if request.Operation != "send" && (request.Text != "" || request.Mode != "") {
		return false
	}
	if request.Operation != "keys" && len(request.Keys) != 0 {
		return false
	}
	if request.Operation != "read" && (request.Coverage != "" || request.MaxBytes != 0) {
		return false
	}
	if request.Ref != "" {
		if request.Name != "" || request.Machine != "" {
			return false
		}
		ref, err := reference.Decode(request.Ref)
		if err != nil {
			return false
		}
		if agentOperation(request.Operation) && ref.Kind != "agent" || !agentOperation(request.Operation) && ref.Kind != "terminal" {
			return false
		}
	}
	switch request.Operation {
	case "list":
		return request.Name == "" && request.Ref == ""
	case "start":
		return request.Name != "" && request.Machine != "" && request.Ref == "" && (request.Kind == LaunchAgent && request.Profile != "" || request.Kind == LaunchTerminal && request.Profile == "" && request.Objective == "") && validDestination(request, true)
	case "info", "shell", "rename", "move", "kill", "enter", "read", "send", "keys", "interrupt", "stop":
		if request.Name == "" && request.Ref == "" {
			return false
		}
	default:
		return false
	}
	switch request.Operation {
	case "rename":
		return request.NewName != ""
	case "move":
		return validDestination(request, false)
	case "send":
		return request.Text != "" && len(request.Text) <= 32768 && utf8.ValidString(request.Text) && (request.Mode == "" || request.Mode == "auto" || request.Mode == "terminal")
	case "read":
		return (request.Coverage == "" || request.Coverage == "recent" || request.Coverage == "visible") && request.MaxBytes >= 0 && request.MaxBytes <= 32768
	case "keys":
		if len(request.Keys) < 1 || len(request.Keys) > 16 {
			return false
		}
		for _, key := range request.Keys {
			switch key {
			case "enter", "escape", "ctrl-c", "up", "down", "left", "right", "tab", "backspace":
			default:
				return false
			}
		}
	}
	return true
}
func validDestination(request Request, optional bool) bool {
	switch request.DestinationKind {
	case "":
		return optional && request.WorkspaceRef == "" && request.NewWorkspace == ""
	case "existing":
		if request.NewWorkspace != "" {
			return false
		}
		ref, err := reference.Decode(request.WorkspaceRef)
		return err == nil && ref.Kind == "workspace"
	case "new":
		return request.WorkspaceRef == "" && (optional || request.NewWorkspace != "")
	default:
		return false
	}
}
func agentOperation(operation string) bool {
	switch operation {
	case "read", "send", "keys", "interrupt", "stop":
		return true
	}
	return false
}
