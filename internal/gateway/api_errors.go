package gateway

import (
	"errors"
	"net/http"

	"github.com/NielsdaWheelz/herdr-mobile/internal/logging"
	"github.com/NielsdaWheelz/herdr-mobile/internal/reference"
	"github.com/NielsdaWheelz/herdr-mobile/internal/sessions"
)

func (gateway *Gateway) writeOperationError(writer http.ResponseWriter, err error) {
	if errors.Is(err, reference.ErrInvalid) {
		failure := errorInvalidRequest
		failure.Dispatch = "not_sent"
		writeError(writer, failure)
		return
	}
	if errors.Is(err, errReferenceMachine) {
		failure := errorMachineIdentityMismatch
		failure.Dispatch = "not_sent"
		writeError(writer, failure)
		return
	}
	var operation *sessions.Error
	if !errors.As(err, &operation) {
		writeError(writer, errorInternal)
		return
	}
	failure := apiError{Code: string(operation.Code), Dispatch: operation.Dispatch, logCode: logging.ErrorCode(operation.Code)}
	switch operation.Code {
	case sessions.ErrorWorkingDirectoryInvalid:
		failure.Status, failure.Message = http.StatusUnprocessableEntity, "Choose a valid working directory."
	case sessions.ErrorWorkingDirectoryUnavailable:
		failure.Code = string(sessions.ErrorWorkingDirectoryInvalid)
		failure.logCode = logging.ErrorWorkingDirectoryInvalid
		failure.Status, failure.Message = http.StatusUnprocessableEntity, "That directory does not exist or cannot be opened."
	case sessions.ErrorProfileUnknown:
		failure.Status, failure.Message = http.StatusUnprocessableEntity, "Choose an available profile."
	case sessions.ErrorTerminalNotFound:
		failure.Status, failure.Message = http.StatusNotFound, "That terminal no longer exists."
	case sessions.ErrorTerminalStale, sessions.ErrorAgentStale, sessions.ErrorWorkspaceStale:
		failure.Status, failure.Message = http.StatusConflict, "The target changed. Refresh and select it again."
	case sessions.ErrorMetadataUnavailable:
		failure.Status, failure.Message = http.StatusBadGateway, "The terminal metadata could not be confirmed."
	case sessions.ErrorNameInvalid:
		failure.Status, failure.Message = http.StatusUnprocessableEntity, "Use a valid terminal name."
	case sessions.ErrorObjectiveInvalid:
		failure.Status, failure.Message = http.StatusUnprocessableEntity, "Use 1–240 characters without terminal controls."
	case sessions.ErrorClosureConfirmationRequired:
		failure.Status, failure.Message = http.StatusConflict, "Herdr requires native confirmation before closing this terminal."
	case sessions.ErrorHerdrUnavailable:
		failure.Status, failure.Message = http.StatusServiceUnavailable, "Herdr is unavailable."
	case sessions.ErrorUpstreamRejected:
		failure.Status, failure.Message = http.StatusBadGateway, "Herdr rejected the operation."
	case sessions.ErrorOutcomeUnknown:
		failure.Status, failure.Message = http.StatusGatewayTimeout, "The operation outcome is unknown. Inspect before retrying."
	default:
		writeError(writer, errorInternal)
		return
	}
	if failure.Dispatch == "" {
		failure.Dispatch = "not_sent"
	}
	if operation.Partial != nil {
		partial := map[string]any{}
		if operation.Partial.Stage != "" {
			partial["stage"] = operation.Partial.Stage
		}
		if operation.Partial.Terminal != nil {
			terminal, mapErr := mapTerminal(gateway.machine, *operation.Partial.Terminal)
			if mapErr != nil {
				writeError(writer, errorInternal)
				return
			}
			partial["terminal"] = terminal
		} else if operation.Partial.TerminalOutcome != "" {
			partial["terminal"] = operation.Partial.TerminalOutcome
		}
		if operation.Partial.AgentOutcome != "" {
			partial["agent"] = operation.Partial.AgentOutcome
		}
		failure.Partial = partial
	}
	writeError(writer, failure)
}
