package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/NielsdaWheelz/skidbladnir/internal/sessions"
	"github.com/NielsdaWheelz/skidbladnir/internal/strictjson"
)

const hostOperationBudget = 10 * time.Second

type createTerminalRequest struct {
	Kind        stringField     `json:"kind"`
	Profile     stringField     `json:"profile"`
	CWD         stringField     `json:"cwd"`
	Name        stringField     `json:"name"`
	Objective   stringField     `json:"objective"`
	Destination json.RawMessage `json:"destination"`
}

type destinationRequest struct {
	Kind         stringField `json:"kind"`
	WorkspaceRef stringField `json:"workspaceRef"`
	Label        stringField `json:"label"`
}

func (gateway *Gateway) destination(encoded json.RawMessage, labelRequired bool) (sessions.Destination, error) {
	if len(encoded) == 0 {
		return sessions.Destination{}, nil
	}
	if bytes.Equal(encoded, []byte("null")) {
		return sessions.Destination{}, strictDestinationError{}
	}
	var input destinationRequest
	if strictjson.Decode(encoded, &input) != nil || !input.Kind.present {
		return sessions.Destination{}, strictDestinationError{}
	}
	switch input.Kind.value {
	case "existing":
		if !input.WorkspaceRef.present || input.Label.present {
			return sessions.Destination{}, strictDestinationError{}
		}
		target, err := parseWorkspaceReference(input.WorkspaceRef.value, gateway.machine)
		if err != nil {
			return sessions.Destination{}, err
		}
		return sessions.Destination{Kind: "existing", WorkspaceTarget: target}, nil
	case "new":
		if input.WorkspaceRef.present || (labelRequired && (!input.Label.present || input.Label.value == "")) {
			return sessions.Destination{}, strictDestinationError{}
		}
		return sessions.Destination{Kind: "new", Label: input.Label.value}, nil
	default:
		return sessions.Destination{}, strictDestinationError{}
	}
}

type strictDestinationError struct{}

func (strictDestinationError) Error() string { return "invalid destination" }

func terminalPath(path string) (ref, operation string, ok bool) {
	rest, found := strings.CutPrefix(path, "/v1/terminals/")
	if !found || rest == "" {
		return "", "", false
	}
	ref, operation, _ = strings.Cut(rest, "/")
	if ref == "" || strings.Contains(operation, "/") {
		return "", "", false
	}
	return ref, operation, true
}

func (gateway *Gateway) serveTerminalAPI(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/v1/terminals" {
		switch request.Method {
		case http.MethodGet:
			if !requireEmptyRequest(request) {
				writeError(writer, errorInvalidRequest)
				return
			}
			ctx, cancel := context.WithTimeout(request.Context(), hostOperationBudget)
			defer cancel()
			inventory, err := gateway.sessions.List(ctx)
			if err != nil {
				gateway.writeOperationError(writer, err)
				return
			}
			for index := range inventory.Terminals {
				gateway.agents.Enrich(ctx, &inventory.Terminals[index])
			}
			result, err := mapInventory(gateway.machine, gateway.machineDTO(), inventory, gateway.sessions.Profiles())
			if err != nil {
				writeError(writer, errorInternal)
				return
			}
			encoded, err := json.Marshal(result)
			if err != nil {
				writeError(writer, errorInternal)
				return
			}
			if len(encoded)+1 > 1<<20 {
				writeError(writer, errorInternal)
				return
			}
			writeJSON(writer, http.StatusOK, result)
		case http.MethodPost:
			gateway.createTerminal(writer, request)
		default:
			writeError(writer, errorInvalidRequest)
		}
		return
	}
	ref, operation, ok := terminalPath(request.URL.Path)
	if !ok {
		writeError(writer, errorInvalidRequest)
		return
	}
	target, err := parseTerminalReference(ref, gateway.machine)
	if err != nil {
		gateway.writeOperationError(writer, err)
		return
	}
	if request.Method == http.MethodGet && operation == "stream" {
		gateway.openTerminal(writer, request, target)
		return
	}
	ctx, cancel := context.WithTimeout(request.Context(), hostOperationBudget)
	defer cancel()
	switch {
	case request.Method == http.MethodGet && operation == "":
		if !requireEmptyRequest(request) {
			writeError(writer, errorInvalidRequest)
			return
		}
		observed, err := gateway.sessions.Info(ctx, target)
		if err != nil {
			gateway.writeOperationError(writer, err)
			return
		}
		gateway.agents.Enrich(ctx, &observed.Terminal)
		gateway.writeObservedTerminal(writer, http.StatusOK, observed)
	case request.Method == http.MethodPost && operation == "shell":
		_, failure := decodeJSON[struct{}](writer, request)
		if failure != nil {
			writeError(writer, *failure)
			return
		}
		created, err := gateway.sessions.Shell(ctx, target)
		gateway.writeCreated(writer, created, err)
	case request.Method == http.MethodPatch && operation == "":
		input, failure := decodeJSON[struct {
			Name stringField `json:"name"`
		}](writer, request)
		if failure != nil {
			writeError(writer, *failure)
			return
		}
		if !input.Name.present {
			writeError(writer, errorInvalidRequest)
			return
		}
		observed, err := gateway.sessions.Rename(ctx, target, input.Name.value)
		if err != nil {
			gateway.writeOperationError(writer, err)
			return
		}
		gateway.writeObservedTerminal(writer, http.StatusOK, observed)
	case request.Method == http.MethodPut && operation == "workspace":
		input, failure := decodeJSON[struct {
			Destination json.RawMessage `json:"destination"`
		}](writer, request)
		if failure != nil {
			writeError(writer, *failure)
			return
		}
		if len(input.Destination) == 0 {
			writeError(writer, errorInvalidRequest)
			return
		}
		destination, err := gateway.destination(input.Destination, true)
		if err != nil {
			gateway.writeDestinationError(writer, err)
			return
		}
		observed, err := gateway.sessions.Move(ctx, target, destination)
		if err != nil {
			gateway.writeOperationError(writer, err)
			return
		}
		gateway.writeObservedTerminal(writer, http.StatusOK, observed)
	case request.Method == http.MethodDelete && operation == "":
		_, failure := decodeJSON[struct{}](writer, request)
		if failure != nil {
			writeError(writer, *failure)
			return
		}
		closed, err := gateway.sessions.Kill(ctx, target)
		if err != nil {
			gateway.writeOperationError(writer, err)
			return
		}
		gateway.closeLiveTerminals(target.TerminalID)
		writeJSON(writer, http.StatusOK, struct {
			Terminal string `json:"terminal"`
			Dispatch string `json:"dispatch"`
		}{closed.Terminal, closed.Dispatch})
	default:
		writeError(writer, errorInvalidRequest)
	}
}

func (gateway *Gateway) createTerminal(writer http.ResponseWriter, request *http.Request) {
	input, failure := decodeJSON[createTerminalRequest](writer, request)
	if failure != nil {
		writeError(writer, *failure)
		return
	}
	if !input.Kind.present || !input.CWD.present || !input.Name.present {
		writeError(writer, errorInvalidRequest)
		return
	}
	if input.Kind.value == "agent" && !input.Profile.present || input.Kind.value == "terminal" && (input.Profile.present || input.Objective.present) {
		writeError(writer, errorInvalidRequest)
		return
	}
	destination, err := gateway.destination(input.Destination, false)
	if err != nil {
		gateway.writeDestinationError(writer, err)
		return
	}
	ctx, cancel := context.WithTimeout(request.Context(), hostOperationBudget)
	defer cancel()
	created, err := gateway.sessions.Create(ctx, sessions.CreateInput{
		Kind:        sessions.LaunchKind(input.Kind.value),
		Profile:     input.Profile.value,
		CWD:         input.CWD.value,
		Name:        input.Name.value,
		Objective:   input.Objective.value,
		Destination: destination,
	})
	gateway.writeCreated(writer, created, err)
}

func (gateway *Gateway) writeDestinationError(writer http.ResponseWriter, err error) {
	if _, invalid := err.(strictDestinationError); invalid {
		writeError(writer, errorInvalidRequest)
		return
	}
	gateway.writeOperationError(writer, err)
}

func (gateway *Gateway) writeCreated(writer http.ResponseWriter, created sessions.Created, err error) {
	if err != nil {
		gateway.writeOperationError(writer, err)
		return
	}
	observedAt, mapErr := projectionInstant(created.ObservedAt)
	if mapErr != nil {
		writeError(writer, errorInternal)
		return
	}
	terminal, mapErr := mapTerminal(gateway.machine, created.Terminal)
	if mapErr != nil {
		writeError(writer, errorInternal)
		return
	}
	writeJSON(writer, http.StatusCreated, struct {
		ObservedAt string       `json:"observedAt"`
		Terminal   terminalWire `json:"terminal"`
		Launch     string       `json:"launch"`
		Dispatch   string       `json:"dispatch"`
	}{observedAt, terminal, created.Launch, created.Dispatch})
}

func (gateway *Gateway) writeObservedTerminal(writer http.ResponseWriter, status int, observed sessions.ObservedTerminal) {
	observedAt, err := projectionInstant(observed.ObservedAt)
	if err != nil {
		writeError(writer, errorInternal)
		return
	}
	terminal, err := mapTerminal(gateway.machine, observed.Terminal)
	if err != nil {
		writeError(writer, errorInternal)
		return
	}
	if observed.Dispatch == "" {
		writeJSON(writer, status, struct {
			ObservedAt string       `json:"observedAt"`
			Terminal   terminalWire `json:"terminal"`
		}{observedAt, terminal})
		return
	}
	writeJSON(writer, status, struct {
		ObservedAt string       `json:"observedAt"`
		Terminal   terminalWire `json:"terminal"`
		Dispatch   string       `json:"dispatch"`
	}{observedAt, terminal, observed.Dispatch})
}
