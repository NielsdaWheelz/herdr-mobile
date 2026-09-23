package sessions

import (
	"time"

	"github.com/NielsdaWheelz/skidbladnir/internal/agentruntime"
	"github.com/NielsdaWheelz/skidbladnir/internal/catalog"
	"github.com/NielsdaWheelz/skidbladnir/internal/herdr"
	"github.com/NielsdaWheelz/skidbladnir/internal/process"
	"github.com/NielsdaWheelz/skidbladnir/internal/workdir"
)

type Config struct {
	Herdr         *herdr.Client
	Workdir       *workdir.Service
	CataloguePath string
	Profiles      []agentruntime.Profile
	MachineHandle string
	Fingerprint   func(process.Observation) (string, error)
}

type TerminalTarget struct {
	TerminalID    string `json:"terminalId"`
	IdentityToken string `json:"identityToken"`
}

type WorkspaceTarget struct {
	WorkspaceID   string `json:"workspaceId"`
	IdentityToken string `json:"identityToken"`
}

type AgentTarget struct {
	TerminalTarget     TerminalTarget        `json:"-"`
	PID                process.PID           `json:"pid"`
	StartIdentity      process.StartIdentity `json:"startIdentity"`
	CommandFingerprint string                `json:"commandFingerprint"`
	Provider           agentruntime.Provider `json:"provider"`
}

type Workspace struct {
	Target WorkspaceTarget
	Label  string
}

type Agent struct {
	Target               AgentTarget
	Provider             agentruntime.Provider
	ProvenRuntimeProfile agentruntime.ProfileKey
	ProviderSession      *agentruntime.ProviderSessionFacts
	Status               agentruntime.Status
	Readiness            string
	Methods              agentruntime.Methods
}

type Terminal struct {
	Target          TerminalTarget
	Name            string
	NativeLabel     string
	Character       catalog.Character
	WorkspaceTarget WorkspaceTarget
	CWD             string
	LaunchProfile   agentruntime.ProfileKey
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

type StopClosure struct {
	Closed         Closed
	AgentExited    bool
	CloseAttempted bool
}

type ResolvedTerminal struct {
	Terminal Terminal
	PaneID   string
}

type ResolvedAgent struct {
	Terminal Terminal
	PaneID   string
	Process  process.Observation
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
	ErrorNameAmbiguous               ErrorCode = "NameAmbiguous"
	ErrorObjectiveInvalid            ErrorCode = "ObjectiveInvalid"
	ErrorReadinessUnconfirmed        ErrorCode = "ReadinessUnconfirmed"
	ErrorMethodUnavailable           ErrorCode = "MethodUnavailable"
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
