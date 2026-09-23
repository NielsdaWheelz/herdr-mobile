package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/NielsdaWheelz/skidbladnir/internal/strictjson"
)

func agentPath(path string) (ref, operation string, ok bool) {
	rest, found := strings.CutPrefix(path, "/v1/agents/")
	if !found || rest == "" {
		return "", "", false
	}
	ref, operation, found = strings.Cut(rest, "/")
	return ref, operation, found && ref != "" && operation != "" && !strings.Contains(operation, "/")
}

func (gateway *Gateway) serveAgentAPI(writer http.ResponseWriter, request *http.Request) {
	ref, operation, ok := agentPath(request.URL.Path)
	if !ok || request.Method != http.MethodPost {
		writeError(writer, errorInvalidRequest)
		return
	}
	target, err := parseAgentReference(ref, gateway.machine)
	if err != nil {
		gateway.writeOperationError(writer, err)
		return
	}
	ctx, cancel := context.WithTimeout(request.Context(), hostOperationBudget)
	defer cancel()
	switch operation {
	case "read":
		input, failure := decodeJSON[struct {
			Coverage stringField     `json:"coverage"`
			MaxBytes json.RawMessage `json:"maxBytes"`
		}](writer, request)
		if failure != nil {
			writeError(writer, *failure)
			return
		}
		maxBytes := 0
		if len(input.MaxBytes) != 0 {
			if bytes.Equal(input.MaxBytes, []byte("null")) || strictjson.Decode(input.MaxBytes, &maxBytes) != nil {
				writeError(writer, errorInvalidRequest)
				return
			}
		}
		result, err := gateway.agents.Read(ctx, target, input.Coverage.value, maxBytes)
		if err != nil {
			gateway.writeOperationError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, result)
	case "send":
		input, failure := decodeJSON[struct {
			Text stringField `json:"text"`
			Mode stringField `json:"mode"`
		}](writer, request)
		if failure != nil {
			writeError(writer, *failure)
			return
		}
		if !input.Text.present {
			writeError(writer, errorInvalidRequest)
			return
		}
		result, err := gateway.agents.Send(ctx, target, input.Text.value, input.Mode.value)
		if err != nil {
			gateway.writeOperationError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, result)
	case "keys":
		input, failure := decodeJSON[struct {
			Keys json.RawMessage `json:"keys"`
		}](writer, request)
		if failure != nil {
			writeError(writer, *failure)
			return
		}
		var keys []string
		if len(input.Keys) == 0 || bytes.Equal(input.Keys, []byte("null")) || strictjson.Decode(input.Keys, &keys) != nil || keys == nil {
			writeError(writer, errorInvalidRequest)
			return
		}
		result, err := gateway.agents.Keys(ctx, target, keys)
		if err != nil {
			gateway.writeOperationError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, result)
	case "interrupt":
		_, failure := decodeJSON[struct{}](writer, request)
		if failure != nil {
			writeError(writer, *failure)
			return
		}
		result, err := gateway.agents.Interrupt(ctx, target)
		if err != nil {
			gateway.writeOperationError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, result)
	case "stop":
		_, failure := decodeJSON[struct{}](writer, request)
		if failure != nil {
			writeError(writer, *failure)
			return
		}
		result, err := gateway.agents.Stop(ctx, target)
		if err != nil {
			gateway.writeOperationError(writer, err)
			return
		}
		gateway.closeLiveTerminals(target.TerminalTarget.TerminalID)
		writeJSON(writer, http.StatusOK, result)
	default:
		writeError(writer, errorInvalidRequest)
	}
}
