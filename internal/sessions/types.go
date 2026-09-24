package sessions

import (
	"time"

	"github.com/NielsdaWheelz/skidbladnir/internal/catalog"
	"github.com/NielsdaWheelz/skidbladnir/internal/herdr"
	"github.com/NielsdaWheelz/skidbladnir/internal/profile"
	"github.com/NielsdaWheelz/skidbladnir/internal/workdir"
)

type Config struct {
	Herdr         *herdr.Client
	Workdir       *workdir.Service
	CataloguePath string
	Profiles      []profile.Profile
	MachineHandle string
}

// TerminalTarget is herdr's terminal_id: stable across pane moves and
// renames, reissued on every herdr restore, and never repeated.
type TerminalTarget struct {
	TerminalID string
}

// WorkspaceTarget is herdr's workspace id. herdr may reissue the number of a
// closed workspace after a restart; a workspace target only places a new tab
// or a moved pane, never input or closure.
type WorkspaceTarget struct {
	WorkspaceID string
}

// AgentTarget is one agent lifetime in one terminal: herdr clears an agent's
// name when that agent exits or another replaces it. An agent that herdr
// detected without a name has an empty Name and is identified by its terminal.
type AgentTarget struct {
	Terminal TerminalTarget
	Name     string
}

type Workspace struct {
	Target WorkspaceTarget
	Label  string
}

type Status struct {
	State  string `json:"state"`
	Source string `json:"source"`
	Reason string `json:"reason,omitempty"`
}

type Agent struct {
	Target    AgentTarget
	Provider  profile.Provider
	Status    Status
	Readiness string
}

type Terminal struct {
	Target          TerminalTarget
	Name            string
	NativeLabel     string
	Character       catalog.Character
	WorkspaceTarget WorkspaceTarget
	CWD             string
	LaunchProfile   profile.Key
	Objective       string
	Agent           *Agent
}

type Inventory struct {
	ObservedAt              time.Time
	Partial                 bool
	UnaddressableTerminals  int
	UnaddressableWorkspaces int
	Workspaces              []Workspace
	Terminals               []Terminal
}

type ObservedTerminal struct {
	ObservedAt time.Time
	Terminal   Terminal
	Dispatch   string
}

type Created struct {
	ObservedAt time.Time
	Terminal   Terminal
	Launch     string
	Dispatch   string
}

type Closed struct {
	Terminal string
	Dispatch string
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

type LaunchKind string

const (
	LaunchAgent    LaunchKind = "agent"
	LaunchTerminal LaunchKind = "terminal"
)

type Destination struct {
	Kind            string
	WorkspaceTarget WorkspaceTarget
	Label           string
}

type CreateInput struct {
	Kind        LaunchKind
	Profile     string
	CWD         string
	Name        string
	Objective   string
	Destination Destination
}

type ErrorCode string

const (
	ErrorWorkingDirectoryInvalid     ErrorCode = "WorkingDirectoryInvalid"
	ErrorWorkingDirectoryUnavailable ErrorCode = "WorkingDirectoryUnavailable"
	ErrorProfileUnknown              ErrorCode = "ProfileUnknown"
	ErrorTerminalNotFound            ErrorCode = "TerminalNotFound"
	ErrorTerminalStale               ErrorCode = "TerminalStale"
	ErrorAgentStale                  ErrorCode = "AgentStale"
	ErrorWorkspaceStale              ErrorCode = "WorkspaceStale"
	ErrorMetadataUnavailable         ErrorCode = "MetadataUnavailable"
	ErrorNameInvalid                 ErrorCode = "NameInvalid"
	ErrorObjectiveInvalid            ErrorCode = "ObjectiveInvalid"
	ErrorClosureConfirmationRequired ErrorCode = "ClosureConfirmationRequired"
	ErrorHerdrUnavailable            ErrorCode = "HerdrUnavailable"
	ErrorUpstreamRejected            ErrorCode = "UpstreamRejected"
	ErrorOutcomeUnknown              ErrorCode = "OutcomeUnknown"
)

type Partial struct {
	Stage           string
	Terminal        *Terminal
	AgentOutcome    string
	TerminalOutcome string
}

type Error struct {
	Code     ErrorCode
	Message  string
	Dispatch string
	Partial  *Partial
}

func (err *Error) Error() string { return err.Message }

func ValidProjectionInstant(value time.Time) bool {
	if value.IsZero() {
		return false
	}
	_, err := value.UTC().MarshalJSON()
	return err == nil
}
