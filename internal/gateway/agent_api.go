package gateway

import (
	"context"
	"net/http"
	"strings"
)

func (gateway *Gateway) serveAgentAPI(writer http.ResponseWriter, request *http.Request) {
	ref, operation, found := strings.Cut(strings.TrimPrefix(request.URL.Path, "/v1/agents/"), "/")
	if !found || ref == "" || request.Method != http.MethodPost || operation != "interrupt" && operation != "stop" {
		writeError(writer, errorInvalidRequest)
		return
	}
	target, err := parseAgentReference(ref, gateway.machine)
	if err != nil {
		gateway.writeOperationError(writer, err)
		return
	}
	if _, failure := decodeJSON[struct{}](writer, request); failure != nil {
		writeError(writer, *failure)
		return
	}
	ctx, cancel := context.WithTimeout(request.Context(), hostOperationBudget)
	defer cancel()
	if operation == "interrupt" {
		result, err := gateway.sessions.Interrupt(ctx, target)
		if err != nil {
			gateway.writeOperationError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, result)
		return
	}
	result, err := gateway.sessions.Stop(ctx, target)
	if err != nil {
		gateway.writeOperationError(writer, err)
		return
	}
	gateway.closeLiveTerminals(target.Terminal.TerminalID)
	writeJSON(writer, http.StatusOK, result)
}
