package logging

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

type Method string

const (
	MethodGet    Method = "GET"
	MethodPost   Method = "POST"
	MethodPut    Method = "PUT"
	MethodPatch  Method = "PATCH"
	MethodDelete Method = "DELETE"
	MethodOther  Method = "OTHER"
)

func (method Method) valid() bool {
	switch method {
	case MethodGet, MethodPost, MethodPut, MethodPatch, MethodDelete, MethodOther:
		return true
	default:
		return false
	}
}

type Route string

const (
	RouteAgentControl      Route = "/v1/agents/{ref}/{operation}"
	RouteTerminals         Route = "/v1/terminals"
	RouteTerminal          Route = "/v1/terminals/{ref}"
	RouteTerminalWorkspace Route = "/v1/terminals/{ref}/workspace"
	RouteTerminalShell     Route = "/v1/terminals/{ref}/shell"
	RouteTerminalStream    Route = "/v1/terminals/{ref}/stream"
	RoutePressure          Route = "/v1/pressure"
	RoutePairingInvites    Route = "/v1/pairing-invites"
	RoutePairings          Route = "/v1/pairings"
	RouteDirectoryListings Route = "/v1/directory-listings"
	RouteUnmatched         Route = "unmatched"
)

func (route Route) valid() bool {
	switch route {
	case RouteAgentControl, RouteTerminals, RouteTerminal, RouteTerminalWorkspace,
		RouteTerminalShell, RouteTerminalStream, RoutePressure, RoutePairingInvites,
		RoutePairings, RouteDirectoryListings, RouteUnmatched:
		return true
	default:
		return false
	}
}

type ErrorCode string

const (
	ErrorNone                        ErrorCode = ""
	ErrorUnauthenticated             ErrorCode = "Unauthenticated"
	ErrorMachineIdentityMismatch     ErrorCode = "MachineIdentityMismatch"
	ErrorInvalidRequest              ErrorCode = "InvalidRequest"
	ErrorRequestTooLarge             ErrorCode = "RequestTooLarge"
	ErrorDirectoryListingUnavailable ErrorCode = "DirectoryListingUnavailable"
	ErrorDirectoryListingTooLarge    ErrorCode = "DirectoryListingTooLarge"
	ErrorPairingInviteRejected       ErrorCode = "PairingInviteRejected"
	ErrorTerminalNotFound            ErrorCode = "TerminalNotFound"
	ErrorTerminalStale               ErrorCode = "TerminalStale"
	ErrorAgentStale                  ErrorCode = "AgentStale"
	ErrorWorkspaceStale              ErrorCode = "WorkspaceStale"
	ErrorMetadataUnavailable         ErrorCode = "MetadataUnavailable"
	ErrorProfileUnknown              ErrorCode = "ProfileUnknown"
	ErrorWorkingDirectoryInvalid     ErrorCode = "WorkingDirectoryInvalid"
	ErrorNameInvalid                 ErrorCode = "NameInvalid"
	ErrorObjectiveInvalid            ErrorCode = "ObjectiveInvalid"
	ErrorClosureConfirmationRequired ErrorCode = "ClosureConfirmationRequired"
	ErrorHerdrUnavailable            ErrorCode = "HerdrUnavailable"
	ErrorUpstreamRejected            ErrorCode = "UpstreamRejected"
	ErrorOutcomeUnknown              ErrorCode = "OutcomeUnknown"
	ErrorInternal                    ErrorCode = "InternalError"
)

func (code ErrorCode) valid() bool {
	switch code {
	case ErrorUnauthenticated, ErrorMachineIdentityMismatch, ErrorInvalidRequest, ErrorRequestTooLarge,
		ErrorDirectoryListingUnavailable, ErrorDirectoryListingTooLarge, ErrorPairingInviteRejected,
		ErrorTerminalNotFound, ErrorTerminalStale, ErrorAgentStale, ErrorWorkspaceStale,
		ErrorMetadataUnavailable, ErrorProfileUnknown, ErrorWorkingDirectoryInvalid,
		ErrorNameInvalid, ErrorObjectiveInvalid,
		ErrorClosureConfirmationRequired, ErrorHerdrUnavailable, ErrorUpstreamRejected,
		ErrorOutcomeUnknown, ErrorInternal:
		return true
	default:
		return false
	}
}

type PressureLevel string

const (
	PressureNormal  PressureLevel = "Normal"
	PressureWarm    PressureLevel = "Warm"
	PressureHot     PressureLevel = "Hot"
	PressureUnknown PressureLevel = "Unknown"
)

func (level PressureLevel) valid() bool {
	switch level {
	case PressureNormal, PressureWarm, PressureHot, PressureUnknown:
		return true
	default:
		return false
	}
}

type PressureReason string

const (
	ReasonMemory    PressureReason = "Memory"
	ReasonDisk      PressureReason = "Disk"
	ReasonLoad      PressureReason = "Load"
	ReasonCPUPSI    PressureReason = "CpuPsi"
	ReasonMemoryPSI PressureReason = "MemoryPsi"
	ReasonIOPSI     PressureReason = "IoPsi"
)

func (reason PressureReason) valid() bool {
	switch reason {
	case ReasonMemory, ReasonDisk, ReasonLoad, ReasonCPUPSI, ReasonMemoryPSI, ReasonIOPSI:
		return true
	default:
		return false
	}
}

// Signal is a shutdown signal the gateway handles, named as the platform names it.
type Signal string

const (
	SignalInterrupt  Signal = "interrupt"
	SignalTerminated Signal = "terminated"
)

type eventKind string

const (
	eventGatewayStarted         eventKind = "Gateway.Started"
	eventGatewayStopping        eventKind = "Gateway.Stopping"
	eventRequestCompleted       eventKind = "Request.Completed"
	eventPressureSampled        eventKind = "Pressure.Sampled"
	eventAuthenticationRejected eventKind = "Authentication.Rejected"
)

type Event struct {
	kind      eventKind
	method    Method
	route     Route
	status    int
	duration  time.Duration
	errorCode ErrorCode
	level     PressureLevel
	reasons   []PressureReason
	signal    Signal
}

func NewGatewayStarted() Event { return Event{kind: eventGatewayStarted} }

func NewGatewayStopping(signal Signal) (Event, error) {
	event := Event{kind: eventGatewayStopping, signal: signal}
	if !event.valid() {
		return Event{}, errors.New("invalid gateway-stopping log event")
	}
	return event, nil
}

func NewRequestCompleted(method Method, route Route, status int, duration time.Duration, errorCode ErrorCode) (Event, error) {
	event := Event{kind: eventRequestCompleted, method: method, route: route, status: status, duration: duration, errorCode: errorCode}
	if !event.valid() {
		return Event{}, errors.New("invalid request log event")
	}
	return event, nil
}

func NewPressureSampled(level PressureLevel, reasons []PressureReason, duration time.Duration) (Event, error) {
	event := Event{kind: eventPressureSampled, level: level, reasons: append([]PressureReason(nil), reasons...), duration: duration}
	if !event.valid() {
		return Event{}, errors.New("invalid pressure-sampled log event")
	}
	return event, nil
}

func NewAuthenticationRejected(route Route) (Event, error) {
	event := Event{kind: eventAuthenticationRejected, route: route}
	if !event.valid() {
		return Event{}, errors.New("invalid authentication-rejected log event")
	}
	return event, nil
}

func (event Event) valid() bool {
	switch event.kind {
	case eventGatewayStarted:
		return true
	case eventGatewayStopping:
		return event.signal == SignalInterrupt || event.signal == SignalTerminated
	case eventRequestCompleted:
		if !event.method.valid() || !event.route.valid() || event.status < 100 || event.status > 599 || event.duration < 0 {
			return false
		}
		return event.status < 400 && event.errorCode == ErrorNone || event.status >= 400 && event.errorCode.valid()
	case eventPressureSampled:
		if !event.level.valid() || event.duration < 0 {
			return false
		}
		seen := make(map[PressureReason]struct{}, len(event.reasons))
		for _, reason := range event.reasons {
			if !reason.valid() {
				return false
			}
			if _, exists := seen[reason]; exists {
				return false
			}
			seen[reason] = struct{}{}
		}
		return true
	case eventAuthenticationRejected:
		return event.route.valid()
	default:
		return false
	}
}

type Logger struct{ output io.Writer }

type WriteError struct{ Err error }

func (err *WriteError) Error() string { return fmt.Sprintf("write structured log: %v", err.Err) }
func (err *WriteError) Unwrap() error { return err.Err }

func New(output io.Writer) Logger { return Logger{output: output} }

func (logger Logger) Write(event Event) error {
	if logger.output == nil || !event.valid() {
		return errors.New("invalid logger write")
	}
	fields := map[string]any{"event.name": event.kind}
	switch event.kind {
	case eventGatewayStarted:
	case eventGatewayStopping:
		fields["signal"] = event.signal
	case eventRequestCompleted:
		fields["http.request.method"] = event.method
		fields["http.route"] = event.route
		fields["http.response.status_code"] = event.status
		fields["skidbladnir.duration.ms"] = event.duration.Milliseconds()
		if event.errorCode != ErrorNone {
			fields["skidbladnir.error.code"] = event.errorCode
		}
	case eventPressureSampled:
		fields["skidbladnir.pressure.level"] = event.level
		reasons := event.reasons
		if reasons == nil {
			reasons = []PressureReason{}
		}
		fields["skidbladnir.pressure.reasons"] = reasons
		fields["skidbladnir.duration.ms"] = event.duration.Milliseconds()
	case eventAuthenticationRejected:
		fields["http.route"] = event.route
	}
	if err := json.NewEncoder(logger.output).Encode(fields); err != nil {
		return &WriteError{Err: err}
	}
	return nil
}
