package herdr

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/NielsdaWheelz/skidbladnir/internal/strictjson"
)

const (
	maxControlLine      = 3 << 20
	maxANSIFrame        = 2 << 20
	controlWriteTimeout = 500 * time.Millisecond
)

var ErrClosed = errors.New("herdr terminal control closed")

type Frame struct {
	Seq           uint64
	Width, Height uint16
	Full          bool
	ANSI          []byte
}

// Control owns one public terminal session control child. Closing it releases
// only this attachment; no Herdr server or worker is signaled.
type Control struct {
	command *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	reader  *bufio.Reader
	mu      sync.Mutex
	closed  bool
}

func (c *Client) OpenControl(ctx context.Context, terminalID string, columns, rows uint16, takeover bool) (*Control, error) {
	if terminalID == "" || columns == 0 || rows == 0 {
		return nil, errors.New("invalid herdr control target or geometry")
	}
	if err := c.Ping(ctx); err != nil {
		var upstream *Error
		if errors.As(err, &upstream) {
			return nil, &Error{Code: upstream.Code, Dispatch: "not_sent"}
		}
		return nil, &Error{Code: "preflight_failed", Dispatch: "not_sent"}
	}
	args := []string{"terminal", "session", "control", terminalID, "--cols", strconv.FormatUint(uint64(columns), 10), "--rows", strconv.FormatUint(uint64(rows), 10)}
	if takeover {
		args = append(args, "--takeover")
	}
	command := exec.Command(c.path, args...)
	command.Env = append(os.Environ(), "HERDR_SOCKET_PATH="+c.socketPath)
	command.Stderr = io.Discard
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, errors.New("herdr control unavailable")
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, errors.New("herdr control unavailable")
	}
	if err := ctx.Err(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, err
	}
	if err := command.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, errors.New("herdr control unavailable")
	}
	control := &Control{command: command, stdin: stdin, stdout: stdout, reader: bufio.NewReader(stdout)}
	if err := ctx.Err(); err != nil {
		_ = control.Close()
		return nil, err
	}
	return control, nil
}

// ReadFrame returns one bounded ANSI frame. A terminal.closed notification is
// an attachment closure, not a typed conflict reason in Herdr v0.9.1.
func (control *Control) ReadFrame() (Frame, error) {
	line, err := readLine(control.reader, maxControlLine)
	if err != nil {
		return Frame{}, fmt.Errorf("herdr control read: %w", err)
	}
	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(line, &envelope); err != nil {
		return Frame{}, errors.New("invalid herdr control event")
	}
	switch envelope.Type {
	case "terminal.closed":
		var closed struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		}
		if err := strictjson.Decode(line, &closed); err != nil {
			return Frame{}, errors.New("invalid herdr close event")
		}
		return Frame{}, ErrClosed
	case "terminal.frame":
		var wire struct {
			Type     string `json:"type"`
			Seq      uint64 `json:"seq"`
			Encoding string `json:"encoding"`
			Width    uint16 `json:"width"`
			Height   uint16 `json:"height"`
			Full     bool   `json:"full"`
			Bytes    string `json:"bytes"`
		}
		if err := strictjson.Decode(line, &wire); err != nil || wire.Seq == 0 || wire.Width == 0 || wire.Height == 0 || wire.Encoding != "ansi" || base64.StdEncoding.DecodedLen(len(wire.Bytes)) > maxANSIFrame+2 {
			return Frame{}, errors.New("invalid herdr frame")
		}
		ansi, err := base64.StdEncoding.Strict().DecodeString(wire.Bytes)
		if err != nil || len(ansi) > maxANSIFrame {
			return Frame{}, errors.New("invalid herdr frame bytes")
		}
		return Frame{Seq: wire.Seq, Width: wire.Width, Height: wire.Height, Full: wire.Full, ANSI: ansi}, nil
	default:
		return Frame{}, errors.New("unexpected herdr control event")
	}
}

func (control *Control) Scroll(source, direction string, lines uint16, column, row *uint16) error {
	if (source != "wheel" && source != "page_key") || (direction != "up" && direction != "down") || lines == 0 {
		return errors.New("invalid herdr scroll")
	}
	return control.write(struct {
		Type      string  `json:"type"`
		Source    string  `json:"source"`
		Direction string  `json:"direction"`
		Lines     uint16  `json:"lines"`
		Column    *uint16 `json:"column,omitempty"`
		Row       *uint16 `json:"row,omitempty"`
		Modifiers uint8   `json:"modifiers"`
	}{"terminal.scroll", source, direction, lines, column, row, 0})
}

func (control *Control) Resize(columns, rows uint16) error {
	if columns == 0 || rows == 0 {
		return errors.New("invalid herdr geometry")
	}
	return control.write(struct {
		Type string `json:"type"`
		Cols uint16 `json:"cols"`
		Rows uint16 `json:"rows"`
	}{"terminal.resize", columns, rows})
}

func (control *Control) write(command any) error {
	encoded, err := json.Marshal(command)
	if err != nil {
		return errors.New("invalid herdr control command")
	}
	control.mu.Lock()
	defer control.mu.Unlock()
	if control.closed {
		return ErrClosed
	}
	if err := control.stdin.(*os.File).SetWriteDeadline(time.Now().Add(controlWriteTimeout)); err != nil {
		return errors.New("herdr control write deadline unavailable")
	}
	line := append(encoded, '\n')
	if count, err := control.stdin.Write(line); err != nil || count != len(line) {
		return errors.New("herdr control write failed")
	}
	return nil
}

func (control *Control) Close() error {
	control.mu.Lock()
	if control.closed {
		control.mu.Unlock()
		return nil
	}
	control.closed = true
	if err := control.stdin.(*os.File).SetWriteDeadline(time.Now().Add(controlWriteTimeout)); err == nil {
		_, _ = control.stdin.Write([]byte("{\"type\":\"terminal.release\"}\n"))
	}
	_ = control.stdin.Close()
	_ = control.stdout.Close()
	control.mu.Unlock()

	done := make(chan error, 1)
	go func() { done <- control.command.Wait() }()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		// A stalled control child can be killed without signaling the worker.
		_ = control.command.Process.Kill()
		<-done
	}
	return nil
}
