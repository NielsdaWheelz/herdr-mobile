// Package terminalclient owns one local tty and exact remote terminal stream.
package terminalclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/NielsdaWheelz/skidbladnir/internal/fleetclient"
	"github.com/NielsdaWheelz/skidbladnir/internal/terminal"
	"github.com/charmbracelet/x/ansi"
	"github.com/coder/websocket"
	"github.com/muesli/cancelreader"
	"golang.org/x/term"
)

var errDimensions = errors.New("terminal size must be 20–1024 columns and 5–512 rows; detached")

func Run(ctx context.Context, client *fleetclient.Client, request fleetclient.Request, input, output *os.File) (result error) {
	if !term.IsTerminal(int(input.Fd())) || !term.IsTerminal(int(output.Fd())) {
		return errors.New("enter requires stdin and stdout ttys")
	}
	columns, rows, err := term.GetSize(int(output.Fd()))
	if err != nil {
		return errors.New("cannot measure terminal")
	}
	if _, err = terminal.EncodeResize(columns, rows); err != nil {
		return errDimensions
	}
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()
	acquisition, cancelAcquisition := context.WithTimeout(ctx, fleetclient.Timeout)
	defer cancelAcquisition()
	connection, failure := client.OpenTerminal(acquisition, request, false)
	if failure != nil {
		return errors.New(failure.Code)
	}
	defer connection.CloseNow()
	original, err := term.MakeRaw(int(input.Fd()))
	if err != nil {
		return errors.New("cannot enter raw terminal mode")
	}
	defer func() {
		_, writeErr := io.WriteString(output, ansi.ResetStyle+ansi.ResetModeCursorKeys+ansi.KeypadNumericMode+ansi.ResetModeFocusEvent+ansi.ShowCursor+ansi.ResetModeMouseNormal+ansi.ResetModeMouseButtonEvent+ansi.ResetModeMouseAnyEvent+ansi.ResetModeMouseExtSgr+ansi.ResetModeBracketedPaste+ansi.ResetModeLightDark+ansi.ResetModifyOtherKeys+ansi.ResetModeLeftRightMargin+ansi.SetTopBottomMargins(0, 0)+ansi.ResetModeAltScreenSaveCursor)
		restoreErr := term.Restore(int(input.Fd()), original)
		if writeErr != nil || restoreErr != nil {
			result = errors.Join(result, errors.New("terminal restoration failed"))
		}
	}()
	if _, err = io.WriteString(output, ansi.SetModeAltScreenSaveCursor+ansi.SetModeBracketedPaste); err != nil {
		return errors.New("cannot set local terminal modes")
	}
	resized := make(chan os.Signal, 1)
	signal.Notify(resized, syscall.SIGWINCH)
	defer signal.Stop(resized)
	return stream(ctx, acquisition, connection, input, output, columns, rows, resized, func() (int, int, error) { return term.GetSize(int(output.Fd())) })
}

type streamEvent struct {
	kind websocket.MessageType
	body []byte
	err  error
}

func stream(ctx, acquisition context.Context, connection *websocket.Conn, input *os.File, output io.Writer, columns, rows int, resized <-chan os.Signal, size func() (int, int, error)) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	connection.SetReadLimit(terminal.MaximumFrameBytes)
	resize, err := terminal.EncodeResize(columns, rows)
	if err != nil {
		return errDimensions
	}
	if err = write(acquisition, connection, resize); err != nil {
		return err
	}
	kind, body, err := connection.Read(acquisition)
	if err != nil {
		return errors.New("terminal control unavailable before first frame")
	}
	frame, err := terminal.ParseServerText(body)
	if err != nil || kind != websocket.MessageText {
		return errors.New("invalid initial terminal frame")
	}
	first, ok := frame.(terminal.Frame)
	if !ok || !first.Full || first.Seq == 0 {
		return errors.New("terminal control unavailable before first full frame")
	}
	if err = render(output, first); err != nil {
		return err
	}
	if acquisition.Err() != nil {
		return errors.New("terminal control unavailable before first frame was applied")
	}
	reader, err := cancelreader.NewReader(input)
	if err != nil {
		return errors.New("cannot acquire cancellable terminal input")
	}
	inputs := make(chan streamEvent)
	outputs := make(chan streamEvent)
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		buffer := make([]byte, 4096)
		for {
			n, err := reader.Read(buffer)
			copyOf := append([]byte(nil), buffer[:n]...)
			select {
			case inputs <- streamEvent{body: copyOf, err: err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()
	go func() {
		defer workers.Done()
		for {
			kind, body, err := connection.Read(ctx)
			select {
			case outputs <- streamEvent{kind: kind, body: body, err: err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()
	defer func() { cancel(); reader.Cancel(); connection.CloseNow(); workers.Wait(); reader.Close() }()
	decoder := inputDecoder{rows: first.Rows}
	var pendingTimer *time.Timer
	var pendingTimeout <-chan time.Time
	defer func() {
		if pendingTimer != nil {
			pendingTimer.Stop()
		}
	}()
	seq := first.Seq
	for {
		select {
		case <-ctx.Done():
			return errors.New("terminal detached; work may continue")
		case <-resized:
			columns, rows, err = size()
			if err != nil {
				return errors.New("cannot measure terminal; detached")
			}
			resize, err = terminal.EncodeResize(columns, rows)
			if err != nil {
				return errDimensions
			}
			decoder.rows = rows
			if err = write(ctx, connection, resize); err != nil {
				return err
			}
		case event := <-outputs:
			if event.err != nil {
				return errors.New("terminal stream lost; work may continue")
			}
			if event.kind != websocket.MessageText {
				return errors.New("invalid terminal response")
			}
			message, err := terminal.ParseServerText(event.body)
			if err != nil {
				return errors.New("invalid terminal response")
			}
			switch value := message.(type) {
			case terminal.Frame:
				if value.Seq != seq+1 {
					return errors.New("terminal frame sequence gap; attachment ended")
				}
				seq = value.Seq
				decoder.rows = value.Rows
				if err = render(output, value); err != nil {
					return err
				}
			case terminal.EndFrame:
				return fmt.Errorf("terminal ended: %s; work may continue", value.Code)
			default:
				return errors.New("invalid terminal response")
			}
		case event := <-inputs:
			if len(event.body) > 0 {
				decoded, err := decoder.Feed(event.body)
				if err != nil {
					return err
				}
				if err = deliver(ctx, connection, output, decoded); err != nil {
					return err
				}
				if decoded.detach {
					return nil
				}
				if len(decoder.pending) > 0 {
					if pendingTimer == nil {
						pendingTimer = time.NewTimer(50 * time.Millisecond)
					} else {
						if !pendingTimer.Stop() {
							select {
							case <-pendingTimer.C:
							default:
							}
						}
						pendingTimer.Reset(50 * time.Millisecond)
					}
					pendingTimeout = pendingTimer.C
				} else {
					pendingTimeout = nil
				}
			}
			if event.err != nil {
				if len(decoder.pending) > 0 || decoder.inPaste {
					return errInput
				}
				return write(ctx, connection, terminal.EncodeDetach())
			}
		case <-pendingTimeout:
			pendingTimeout = nil
			decoded, err := decoder.FlushEscape()
			if err != nil {
				return err
			}
			if err = deliver(ctx, connection, output, decoded); err != nil {
				return err
			}
			if decoded.detach {
				return nil
			}
		}
	}
}
func render(output io.Writer, frame terminal.Frame) error {
	if frame.Full {
		if _, err := io.WriteString(output, ansi.ResetStyle+ansi.EraseEntireDisplay+ansi.CursorHomePosition); err != nil {
			return errors.New("terminal output unavailable")
		}
	}
	remaining := frame.ANSI
	for len(remaining) > 0 {
		n, err := output.Write(remaining)
		if err != nil || n <= 0 {
			return errors.New("terminal output unavailable")
		}
		remaining = remaining[n:]
	}
	return nil
}
func deliver(ctx context.Context, connection *websocket.Conn, output io.Writer, decoded decodedInput) error {
	if decoded.notice != "" {
		if _, err := fmt.Fprintf(output, "\r\n%s\r\n", decoded.notice); err != nil {
			return errors.New("terminal output unavailable")
		}
	}
	for _, frame := range decoded.frames {
		var encoded []byte
		var err error
		switch value := frame.(type) {
		case terminal.TextFrame:
			encoded, err = terminal.EncodeText(value.Text)
		case terminal.PasteFrame:
			encoded, err = terminal.EncodePaste(value.Text)
		case terminal.KeyFrame:
			encoded, err = terminal.EncodeKey(value.Key, value.Modifiers)
		case terminal.ScrollFrame:
			encoded, err = terminal.EncodeScroll(value.Source, value.Direction, value.Lines, value.Column, value.Row)
		default:
			return errInput
		}
		if err != nil {
			return errInput
		}
		if err = write(ctx, connection, encoded); err != nil {
			return err
		}
	}
	if decoded.detach {
		return write(ctx, connection, terminal.EncodeDetach())
	}
	return nil
}
func write(parent context.Context, connection *websocket.Conn, body []byte) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	if connection.Write(ctx, websocket.MessageText, body) != nil {
		return errors.New("terminal input delivery unknown; not repeated")
	}
	return nil
}
