// Package fleetclient owns direct peer routing and exact resource selection.
package fleetclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/NielsdaWheelz/skidbladnir/internal/reference"
	"github.com/NielsdaWheelz/skidbladnir/internal/strictjson"
	"github.com/coder/websocket"
)

const (
	MaximumInputBytes     = 256 * 1024
	MaximumControlBytes   = 256 * 1024
	MaximumErrorBytes     = 64 * 1024
	MaximumInventoryBytes = 1024 * 1024
	Timeout               = 15 * time.Second
)

type Failure struct {
	Code     string          `json:"code"`
	Message  string          `json:"message,omitempty"`
	Dispatch string          `json:"dispatch,omitempty"`
	Partial  json.RawMessage `json:"-"`
}
type Result struct {
	Label      string          `json:"label,omitempty"`
	Machine    string          `json:"machine,omitempty"`
	OK         bool            `json:"ok"`
	Value      any             `json:"result,omitempty"`
	Error      *Failure        `json:"error,omitempty"`
	Partial    json.RawMessage `json:"partial,omitempty"`
	Candidates []string        `json:"-"`
}

func Failed(code, dispatch string) Result {
	return Result{Error: &Failure{Code: code, Message: localMessage(code), Dispatch: dispatch}}
}
func localMessage(code string) string {
	switch code {
	case "inventory_incomplete":
		return "A peer or resource is unavailable; select a machine or exact reference."
	case "name_ambiguous":
		return "Several terminals have that name; select an exact reference."
	case "name_not_found":
		return "No terminal has that name."
	case "OutcomeUnknown":
		return "Delivery may have occurred; inspect before retrying."
	case "protocol_error":
		return "The gateway response is invalid."
	case "configuration_invalid":
		return "The private peer configuration is unavailable or invalid."
	default:
		return "The operation could not be completed."
	}
}
func success(value any) Result { return Result{OK: true, Value: value} }

type Client struct {
	peers []peer
	http  *http.Client
}

func (client *Client) Execute(ctx context.Context, request Request) Result {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	if !request.Valid() {
		return Failed("invalid_input", "not_sent")
	}
	if request.Operation == "list" {
		return client.list(ctx, request.Machine)
	}
	if request.Operation == "enter" {
		return Failed("invalid_input", "not_sent")
	}
	var selected peer
	var encodedRef string
	if request.Operation == "start" {
		var ok bool
		selected, ok = client.peerByLabel(request.Machine)
		if !ok {
			return Failed("machine_unknown", "not_sent")
		}
	} else {
		var failed *Result
		selected, encodedRef, failed = client.resolve(ctx, request)
		if failed != nil {
			return *failed
		}
	}
	destination := func() (any, bool) {
		switch request.DestinationKind {
		case "":
			return nil, true
		case "new":
			return struct {
				Kind  string `json:"kind"`
				Label string `json:"label,omitempty"`
			}{"new", request.NewWorkspace}, true
		case "existing":
			value, err := reference.Decode(request.WorkspaceRef)
			if err != nil || value.Kind != "workspace" || value.Machine != selected.Machine {
				return nil, false
			}
			return struct {
				Kind         string `json:"kind"`
				WorkspaceRef string `json:"workspaceRef"`
			}{"existing", request.WorkspaceRef}, true
		default:
			return nil, false
		}
	}
	var method, path string
	var body any
	switch request.Operation {
	case "start":
		dest, ok := destination()
		if !ok {
			return Failed("invalid_input", "not_sent")
		}
		method, path = http.MethodPost, "/v1/terminals"
		body = struct {
			Kind        LaunchKind `json:"kind"`
			Profile     string     `json:"profile,omitempty"`
			CWD         string     `json:"cwd"`
			Name        string     `json:"name"`
			Objective   string     `json:"objective,omitempty"`
			Destination any        `json:"destination,omitempty"`
		}{request.Kind, request.Profile, request.CWD, request.Name, request.Objective, dest}
	case "info":
		method, path = http.MethodGet, "/v1/terminals/"+encodedRef
	case "shell":
		method, path = http.MethodPost, "/v1/terminals/"+encodedRef+"/shell"
		body = struct{}{}
	case "rename":
		method, path = http.MethodPatch, "/v1/terminals/"+encodedRef
		body = struct {
			Name string `json:"name"`
		}{request.NewName}
	case "move":
		dest, ok := destination()
		if !ok {
			return Failed("invalid_input", "not_sent")
		}
		method, path = http.MethodPut, "/v1/terminals/"+encodedRef+"/workspace"
		body = struct {
			Destination any `json:"destination"`
		}{dest}
	case "kill":
		method, path = http.MethodDelete, "/v1/terminals/"+encodedRef
		body = struct{}{}
	case "read":
		method, path = http.MethodPost, "/v1/agents/"+encodedRef+"/read"
		body = struct {
			Coverage string `json:"coverage,omitempty"`
			MaxBytes int    `json:"maxBytes,omitempty"`
		}{request.Coverage, request.MaxBytes}
	case "send":
		method, path = http.MethodPost, "/v1/agents/"+encodedRef+"/send"
		mode := request.Mode
		if mode == "" {
			mode = "auto"
		}
		body = struct {
			Text string `json:"text"`
			Mode string `json:"mode"`
		}{request.Text, mode}
	case "keys":
		method, path = http.MethodPost, "/v1/agents/"+encodedRef+"/keys"
		body = struct {
			Keys []string `json:"keys"`
		}{request.Keys}
	case "interrupt", "stop":
		method, path = http.MethodPost, "/v1/agents/"+encodedRef+"/"+request.Operation
		body = struct{}{}
	default:
		return Failed("invalid_input", "not_sent")
	}
	var data []byte
	if body != nil {
		data, _ = json.Marshal(body)
		if len(data) > MaximumInputBytes {
			return Failed("input_limit", "not_sent")
		}
	}
	return client.call(ctx, selected, request.Operation, method, path, data)
}
func (client *Client) list(ctx context.Context, label string) Result {
	peers := client.peers
	if label != "" {
		selected, ok := client.peerByLabel(label)
		if !ok {
			return Failed("machine_unknown", "not_sent")
		}
		peers = []peer{selected}
	}
	rows := make([]Peer, len(peers))
	var pending sync.WaitGroup
	for index, selected := range peers {
		pending.Go(func() {
			result := client.call(ctx, selected, "list", http.MethodGet, "/v1/terminals", nil)
			row := Peer{Label: selected.Label, Machine: selected.Machine, Error: result.Error}
			if result.OK {
				row = result.Value.(Peer)
			}
			rows[index] = row
		})
	}
	pending.Wait()
	inventory := Inventory{Peers: rows}
	for _, row := range rows {
		if !row.OK || row.Partial {
			inventory.Partial = true
		}
	}
	return success(inventory)
}
func (client *Client) resolve(ctx context.Context, request Request) (peer, string, *Result) {
	fail := func(code string) (peer, string, *Result) {
		r := Failed(code, "not_sent")
		return peer{}, "", &r
	}
	if request.Ref != "" {
		value, err := reference.Decode(request.Ref)
		if err != nil {
			return fail("invalid_input")
		}
		selected, ok := client.peerByMachine(value.Machine)
		if !ok {
			return fail("machine_unknown")
		}
		return selected, request.Ref, nil
	}
	inventory := client.list(ctx, request.Machine)
	if !inventory.OK {
		return peer{}, "", &inventory
	}
	listed := inventory.Value.(Inventory)
	if listed.Partial {
		return fail("inventory_incomplete")
	}
	type match struct {
		peer Peer
		ref  string
	}
	var matches []match
	for _, p := range listed.Peers {
		for _, t := range p.Terminals {
			if t.Name == request.Name {
				ref := t.Ref
				if agentOperation(request.Operation) {
					if t.Agent == nil {
						continue
					}
					ref = t.Agent.Ref
				}
				matches = append(matches, match{p, ref})
			}
		}
	}
	if len(matches) == 0 {
		return fail("name_not_found")
	}
	if len(matches) > 1 {
		r := Failed("name_ambiguous", "not_sent")
		for _, m := range matches {
			r.Candidates = append(r.Candidates, m.peer.Label+" / "+request.Name+" / "+m.ref)
		}
		return peer{}, "", &r
	}
	m := matches[0]
	selected, _ := client.peerByMachine(m.peer.Machine)
	return selected, m.ref, nil
}
func (client *Client) OpenTerminal(ctx context.Context, request Request, takeover bool) (*websocket.Conn, *Failure) {
	request.Operation = "enter"
	if !request.Valid() {
		return nil, &Failure{Code: "invalid_input", Dispatch: "not_sent"}
	}
	selected, encoded, failed := client.resolve(ctx, request)
	if failed != nil {
		return nil, failed.Error
	}
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+selected.Bearer)
	headers.Set("Skidbladnir-Machine", selected.Machine)
	if takeover {
		headers.Set("Skidbladnir-Terminal-Takeover", "true")
	} else {
		headers.Set("Skidbladnir-Terminal-Takeover", "false")
	}
	connection, response, err := websocket.Dial(ctx, strings.Replace(selected.Origin, "https://", "wss://", 1)+"/v1/terminals/"+encoded+"/stream", &websocket.DialOptions{HTTPClient: client.http, HTTPHeader: headers, CompressionMode: websocket.CompressionDisabled})
	if err == nil {
		return connection, nil
	}
	if response != nil && response.Body != nil {
		defer response.Body.Close()
		data, readErr := io.ReadAll(io.LimitReader(response.Body, MaximumErrorBytes+1))
		if readErr == nil && len(data) <= MaximumErrorBytes {
			if failure := decodeFailure(data); failure != nil {
				return nil, failure
			}
		}
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, &Failure{Code: "control_unavailable", Message: "Terminal acquisition timed out.", Dispatch: "not_sent"}
	}
	return nil, &Failure{Code: "stream_lost", Dispatch: "not_sent"}
}
func (client *Client) call(ctx context.Context, target peer, operation, method, path string, body []byte) (answer Result) {
	defer func() {
		if !answer.OK {
			answer.Label = target.Label
			answer.Machine = target.Machine
		}
	}()
	dispatch := "not_sent"
	writes := method != http.MethodGet
	if ctx.Err() != nil {
		return Failed("HerdrUnavailable", dispatch)
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	request, err := http.NewRequestWithContext(ctx, method, target.Origin+path, reader)
	if err != nil {
		return Failed("invalid_input", dispatch)
	}
	request.Header.Set("Authorization", "Bearer "+target.Bearer)
	request.Header.Set("Skidbladnir-Machine", target.Machine)
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if writes {
		dispatch = "unknown"
	}
	response, err := client.http.Do(request)
	if err != nil {
		return Failed("OutcomeUnknown", dispatch)
	}
	defer response.Body.Close()
	expected := http.StatusOK
	if operation == "start" || operation == "shell" {
		expected = http.StatusCreated
	}
	limit := MaximumErrorBytes
	if response.StatusCode == expected {
		limit = MaximumControlBytes
		if operation == "list" {
			limit = MaximumInventoryBytes
		}
	}
	encoded, err := io.ReadAll(io.LimitReader(response.Body, int64(limit+1)))
	if err != nil {
		return Failed("OutcomeUnknown", dispatch)
	}
	if len(encoded) > limit {
		return Failed("output_limit", dispatch)
	}
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return Failed("protocol_error", dispatch)
	}
	if response.StatusCode != expected {
		failure := decodeFailure(encoded)
		if failure == nil {
			return Failed("protocol_error", dispatch)
		}
		return Result{Error: failure, Partial: failure.Partial}
	}
	value, err := decodeSuccess(operation, encoded, target)
	if err != nil {
		return Failed("protocol_error", dispatch)
	}
	return success(value)
}
func decodeFailure(encoded []byte) *Failure {
	var wire *struct {
		Code     string          `json:"code"`
		Message  string          `json:"message"`
		Dispatch string          `json:"dispatch,omitempty"`
		Partial  json.RawMessage `json:"partial,omitempty"`
	}
	if strictjson.Decode(encoded, &wire) != nil || wire == nil || wire.Code == "" || wire.Message == "" {
		return nil
	}
	if wire.Dispatch != "" && wire.Dispatch != "not_sent" && wire.Dispatch != "sent" && wire.Dispatch != "unknown" {
		return nil
	}
	return &Failure{Code: wire.Code, Message: wire.Message, Dispatch: wire.Dispatch, Partial: wire.Partial}
}
func (client *Client) peerByLabel(label string) (peer, bool) {
	for _, p := range client.peers {
		if p.Label == label {
			return p, true
		}
	}
	return peer{}, false
}
func (client *Client) peerByMachine(handle string) (peer, bool) {
	for _, p := range client.peers {
		if p.Machine == handle {
			return p, true
		}
	}
	return peer{}, false
}
func (result Result) Encode(operation string) ([]byte, error) {
	encoded, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	if operation != "list" && len(encoded) > MaximumControlBytes {
		return nil, errors.New("output limit")
	}
	return append(encoded, '\n'), nil
}
func (result Result) ExitCode(operation string) int {
	if !result.OK {
		return 1
	}
	if operation == "list" && result.Value.(Inventory).Partial {
		return 1
	}
	if operation == "send" || operation == "keys" || operation == "interrupt" {
		if value, ok := result.Value.(map[string]json.RawMessage); ok && string(value["outcome"]) == `"unknown"` {
			return 1
		}
	}
	return 0
}
