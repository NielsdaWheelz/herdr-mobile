package sessions

import (
	"time"

	"github.com/NielsdaWheelz/herdr-mobile/internal/catalog"
	"github.com/NielsdaWheelz/herdr-mobile/internal/herdr"
	"github.com/NielsdaWheelz/herdr-mobile/internal/profile"
	"github.com/NielsdaWheelz/herdr-mobile/internal/workdir"
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

// AgentTarget is one named agent lifetime in one terminal: herdr clears an
// agent's name when that agent exits or another replaces it, and names are
// unique among a server's live agents. Name is never empty; an agent herdr
// detected without a name has no target.
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

// Agent is herdr's agent in a terminal. Name is herdr's agent name, empty when
// herdr detected the agent without one; only a named agent can be addressed.
type Agent struct {
	Name      string
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
