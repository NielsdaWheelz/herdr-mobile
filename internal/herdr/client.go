// Package herdr speaks the public Herdr v0.9.1 Unix JSON API and terminal
// session control command. It does not own the Herdr server or its workers.
package herdr

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/NielsdaWheelz/herdr-mobile/internal/strictjson"
)

const (
	apiProtocol    = 22
	maxAPIRequest  = 1 << 20
	maxAPIResponse = 1 << 20
	requestTimeout = 10 * time.Second
	versionTimeout = 5 * time.Second
)

type Client struct {
	path          string
	socketPath    string
	testedVersion string
}

// Error reports a public error code and whether an action reached Herdr.
// It never includes Herdr's free-text diagnostic, which can contain host data.
type Error struct {
	Code     string
	Dispatch string // not_sent, sent, or unknown
}

func (e *Error) Error() string { return "herdr " + e.Code + " (" + e.Dispatch + ")" }

// New verifies the configured binary once. Each subsequent operation checks
// the running server independently through its public ping operation.
func New(ctx context.Context, path, socketPath, testedVersion string) (*Client, error) {
	if !filepath.IsAbs(path) || !filepath.IsAbs(socketPath) || testedVersion != "herdr 0.9.1" {
		return nil, errors.New("invalid herdr configuration")
	}
	ctx, cancel := context.WithTimeout(ctx, versionTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, path, "--version")
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = io.Discard
	if err := command.Run(); err != nil || strings.TrimSpace(output.String()) != testedVersion {
		return nil, errors.New("configured herdr binary version does not match tested version")
	}
	return &Client{path: path, socketPath: socketPath, testedVersion: strings.TrimPrefix(testedVersion, "herdr ")}, nil
}

// Call submits exactly one operation. Herdr closes each API socket after one
// request, so the version ping necessarily uses a separate socket immediately
// before the operation. Socket replacement between the two is a residual race
// in the public protocol; no request is retried.
func (c *Client) Call(ctx context.Context, method string, params any, result any) error {
	if method == "" || strings.IndexFunc(method, func(r rune) bool {
		return r <= ' ' || r > '~'
	}) >= 0 {
		return &Error{Code: "invalid_method", Dispatch: "not_sent"}
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	if err := c.Ping(ctx); err != nil {
		var upstream *Error
		if errors.As(err, &upstream) {
			return &Error{Code: upstream.Code, Dispatch: "not_sent"}
		}
		return &Error{Code: "preflight_failed", Dispatch: "not_sent"}
	}
	return c.request(ctx, method, params, result)
}

// Ping verifies the configured local server's exact tested API version.
func (c *Client) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	var pong struct {
		Type         string          `json:"type"`
		Version      string          `json:"version"`
		Protocol     uint32          `json:"protocol"`
		Capabilities json.RawMessage `json:"capabilities"`
	}
	if err := c.request(ctx, "ping", struct{}{}, &pong); err != nil {
		return err
	}
	if pong.Type != "pong" || pong.Version != c.testedVersion || pong.Protocol != apiProtocol {
		return &Error{Code: "server_mismatch", Dispatch: "not_sent"}
	}
	return nil
}

func (c *Client) request(ctx context.Context, method string, params any, result any) error {
	if params == nil {
		params = struct{}{}
	}
	encoded, err := json.Marshal(struct {
		ID     string `json:"id"`
		Method string `json:"method"`
		Params any    `json:"params"`
	}{ID: "herdr-mobile", Method: method, Params: params})
	if err != nil {
		return &Error{Code: "invalid_request", Dispatch: "not_sent"}
	}
	if len(encoded)+1 > maxAPIRequest {
		return &Error{Code: "request_too_large", Dispatch: "not_sent"}
	}
	connection, err := (&net.Dialer{}).DialContext(ctx, "unix", c.socketPath)
	if err != nil {
		return &Error{Code: "unavailable", Dispatch: "not_sent"}
	}
	defer connection.Close()
	stop := context.AfterFunc(ctx, func() { connection.Close() })
	defer stop()
	if deadline, ok := ctx.Deadline(); ok {
		if err := connection.SetDeadline(deadline); err != nil {
			return &Error{Code: "deadline_failed", Dispatch: "not_sent"}
		}
	}
	for pending := append(encoded, '\n'); len(pending) > 0; {
		written, err := connection.Write(pending)
		if err != nil || written == 0 {
			return &Error{Code: "write_failed", Dispatch: "unknown"}
		}
		pending = pending[written:]
	}
	line, err := readLine(bufio.NewReader(connection), maxAPIResponse)
	if err != nil {
		return &Error{Code: "response_lost", Dispatch: "unknown"}
	}
	var envelope struct {
		ID     string          `json:"id"`
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := strictjson.Decode(line, &envelope); err != nil || envelope.ID != "herdr-mobile" {
		return &Error{Code: "invalid_response", Dispatch: "unknown"}
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(line, &members); err != nil || len(members) != 2 {
		return &Error{Code: "invalid_response", Dispatch: "unknown"}
	}
	_, hasResult := members["result"]
	_, hasError := members["error"]
	resultDocument := bytes.TrimSpace(envelope.Result)
	if hasResult == hasError || (hasResult && (len(resultDocument) == 0 || resultDocument[0] != '{')) || (hasError && envelope.Error == nil) {
		return &Error{Code: "invalid_response", Dispatch: "unknown"}
	}
	if envelope.Error != nil {
		code := envelope.Error.Code
		if code == "" || len(code) > 64 || strings.IndexFunc(code, func(r rune) bool {
			return r != '_' && (r < 'a' || r > 'z') && (r < '0' || r > '9')
		}) >= 0 {
			return &Error{Code: "invalid_response", Dispatch: "unknown"}
		}
		return &Error{Code: code, Dispatch: "sent"}
	}
	var tag struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(envelope.Result, &tag); err != nil || tag.Type == "" {
		return &Error{Code: "invalid_response", Dispatch: "unknown"}
	}
	if result != nil {
		// The envelope has already rejected duplicate members throughout the
		// document. Projection may omit unrelated upstream result fields.
		if err := json.Unmarshal(envelope.Result, result); err != nil {
			return &Error{Code: "invalid_result", Dispatch: "unknown"}
		}
	}
	return nil
}

// readLine bounds a complete newline-delimited document without consuming
// bytes from the next document.
func readLine(reader *bufio.Reader, limit int) ([]byte, error) {
	var line []byte
	for {
		part, err := reader.ReadSlice('\n')
		if len(line)+len(part) > limit+1 {
			return nil, errors.New("herdr line exceeds limit")
		}
		line = append(line, part...)
		if err == nil {
			if len(line) == 1 {
				return nil, errors.New("empty herdr line")
			}
			return line[:len(line)-1], nil
		}
		if !errors.Is(err, bufio.ErrBufferFull) {
			return nil, err
		}
	}
}
