package gateway

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/NielsdaWheelz/herdr-mobile/internal/herdr"
	"github.com/NielsdaWheelz/herdr-mobile/internal/logging"
	"github.com/NielsdaWheelz/herdr-mobile/internal/sessions"
	"github.com/NielsdaWheelz/herdr-mobile/internal/terminal"
	"github.com/coder/websocket"
)

const (
	terminalAcquisitionTimeout = 10 * time.Second
	terminalWriteTimeout       = 5 * time.Second
	terminalPingInterval       = 2 * time.Second
	terminalPongTimeout        = 6 * time.Second
	terminalBearerInterval     = 2 * time.Second
	terminalDetachGrace        = 400 * time.Millisecond
	terminalInputQueueFrames   = 64
)

type liveTerminal struct {
	terminalID string
	cancel     context.CancelFunc
	done       chan struct{}
}

type controlFrame struct {
	frame herdr.Frame
	err   error
}

type clientFrame struct {
	frame terminal.ClientFrame
	ack   chan struct{}
}

func (gateway *Gateway) openTerminal(writer http.ResponseWriter, request *http.Request, target sessions.TerminalTarget) {
	if !requireEmptyRequest(request) || gateway.herdr == nil {
		writeError(writer, errorInvalidRequest)
		return
	}
	takeover := false
	takeoverValues := request.Header.Values("Herdr-Mobile-Terminal-Takeover")
	if len(takeoverValues) != 1 {
		writeError(writer, errorInvalidRequest)
		return
	}
	switch takeoverValues[0] {
	case "false":
	case "true":
		takeover = true
	default:
		writeError(writer, errorInvalidRequest)
		return
	}
	ctx, cancel := context.WithTimeout(request.Context(), hostOperationBudget)
	err := gateway.sessions.CheckTerminal(ctx, target)
	cancel()
	if err != nil {
		gateway.writeOperationError(writer, err)
		return
	}
	connection, err := websocket.Accept(writer, request, &websocket.AcceptOptions{CompressionMode: websocket.CompressionDisabled})
	if err != nil {
		if tracked, ok := writer.(interface{ setErrorCode(logging.ErrorCode) }); ok {
			tracked.setErrorCode(logging.ErrorInvalidRequest)
		}
		return
	}
	connection.SetReadLimit(terminal.MaximumClientFrameBytes)
	attempt, cancel := context.WithCancel(request.Context())
	acquisitionTimer := time.AfterFunc(terminalAcquisitionTimeout, cancel)
	registered, ok := gateway.registerLiveTerminal(target.TerminalID, cancel)
	if !ok {
		cancel()
		writeTerminalEnd(connection, terminal.EndControlUnavailable)
		_ = connection.CloseNow()
		return
	}
	defer func() {
		acquisitionTimer.Stop()
		cancel()
		_ = connection.CloseNow()
		gateway.unregisterLiveTerminal(registered)
	}()
	gateway.runTerminal(attempt, cancel, connection, target, takeover, request.Header.Get("Authorization"), acquisitionTimer.Stop)
}

func (gateway *Gateway) runTerminal(ctx context.Context, cancelAttempt context.CancelFunc, connection *websocket.Conn, target sessions.TerminalTarget, takeover bool, authorization string, acquired func() bool) {
	acquisition, cancelAcquisition := context.WithTimeout(ctx, terminalAcquisitionTimeout)
	defer cancelAcquisition()
	var endReason atomic.Uint32
	writeCanceledEnd := func(admitted bool) {
		switch endReason.Load() {
		case 1:
			writeTerminalEnd(connection, terminal.EndDetached)
		case 2:
			writeTerminalEnd(connection, terminal.EndProtocolError)
		case 3:
			writeTerminalEnd(connection, terminal.EndStreamLost)
		default:
			if !admitted {
				writeTerminalEnd(connection, terminal.EndControlUnavailable)
			}
		}
	}
	messageType, payload, err := connection.Read(acquisition)
	if err != nil || messageType != websocket.MessageText {
		writeTerminalEnd(connection, terminal.EndControlUnavailable)
		return
	}
	initial, err := terminal.ParseClientText(payload)
	resize, ok := initial.(terminal.ResizeFrame)
	if err != nil || !ok {
		writeTerminalEnd(connection, terminal.EndProtocolError)
		return
	}
	inputs := make(chan clientFrame, terminalInputQueueFrames)
	go func() {
		for {
			kind, payload, err := connection.Read(ctx)
			if err != nil {
				cancelAttempt()
				return
			}
			if kind != websocket.MessageText {
				endReason.CompareAndSwap(0, 2)
				cancelAttempt()
				return
			}
			frame, err := terminal.ParseClientText(payload)
			if err != nil {
				endReason.CompareAndSwap(0, 2)
				cancelAttempt()
				return
			}
			if _, detached := frame.(terminal.DetachFrame); detached {
				endReason.CompareAndSwap(0, 1)
				ack := make(chan struct{})
				timer := time.NewTimer(terminalDetachGrace)
				select {
				case inputs <- clientFrame{frame: frame, ack: ack}:
				case <-timer.C:
					cancelAttempt()
					return
				case <-ctx.Done():
					timer.Stop()
					return
				}
				select {
				case <-ack:
				case <-timer.C:
					cancelAttempt()
				case <-ctx.Done():
				}
				timer.Stop()
				return
			}
			select {
			case inputs <- clientFrame{frame: frame}:
			default:
				endReason.CompareAndSwap(0, 2)
				cancelAttempt()
				return
			}
		}
	}()
	if err := gateway.sessions.CheckTerminal(ctx, target); err != nil {
		writeCanceledEnd(false)
		return
	}
	control, err := gateway.herdr.OpenControl(ctx, target.TerminalID, uint16(resize.Columns), uint16(resize.Rows), takeover)
	if err != nil {
		writeCanceledEnd(false)
		return
	}
	defer control.Close()
	if err := gateway.sessions.CheckTerminal(ctx, target); err != nil {
		writeCanceledEnd(false)
		return
	}

	frames := make(chan controlFrame, 1)
	go func() {
		for {
			frame, err := control.ReadFrame()
			select {
			case frames <- controlFrame{frame, err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()
	go watchTerminalPing(ctx, connection, &endReason, cancelAttempt)
	go gateway.watchTerminalBearer(ctx, authorization, &endReason, cancelAttempt)

	var lastSeq uint64
	admitted := false
	acquisitionDone := acquisition.Done()
	for {
		select {
		case <-acquisitionDone:
			writeCanceledEnd(false)
			return
		case <-ctx.Done():
			writeCanceledEnd(admitted)
			return
		case incoming := <-frames:
			if incoming.err != nil {
				if ctx.Err() != nil {
					writeCanceledEnd(admitted)
				} else {
					writeTerminalEnd(connection, terminal.EndStreamLost)
				}
				return
			}
			frame := incoming.frame
			if !admitted && !frame.Full || admitted && (frame.Seq != lastSeq+1 || lastSeq == ^uint64(0)) {
				writeTerminalEnd(connection, terminal.EndProtocolError)
				return
			}
			encoded, err := terminal.EncodeFrame(frame.Seq, int(frame.Width), int(frame.Height), frame.Full, frame.ANSI)
			if err != nil {
				writeTerminalEnd(connection, terminal.EndProtocolError)
				return
			}
			writeContext, cancel := context.WithTimeout(ctx, terminalWriteTimeout)
			err = connection.Write(writeContext, websocket.MessageText, encoded)
			cancel()
			if err != nil {
				if ctx.Err() != nil {
					writeCanceledEnd(admitted)
				}
				return
			}
			lastSeq, admitted = frame.Seq, true
			if acquisitionDone != nil {
				acquisitionDone = nil
				acquired()
			}
		case incoming := <-inputs:
			if _, detached := incoming.frame.(terminal.DetachFrame); detached {
				close(incoming.ack)
				writeTerminalEnd(connection, terminal.EndDetached)
				return
			}
			if !admitted {
				if _, ok := incoming.frame.(terminal.ResizeFrame); !ok {
					writeTerminalEnd(connection, terminal.EndProtocolError)
					return
				}
			}
			inputContext, cancelInput := context.WithTimeout(ctx, hostOperationBudget)
			err := gateway.dispatchTerminalInput(inputContext, target, control, incoming.frame)
			cancelInput()
			if err != nil {
				if ctx.Err() != nil {
					writeCanceledEnd(admitted)
				} else {
					writeTerminalEnd(connection, terminal.EndStreamLost)
				}
				return
			}
		}
	}
}

// dispatchTerminalInput writes text, paste and keys through sessions, which
// re-reads the terminal's pane first; scroll and resize go to this
// attachment's control child, which herdr binds to the terminal itself.
func (gateway *Gateway) dispatchTerminalInput(ctx context.Context, target sessions.TerminalTarget, control *herdr.Control, frame terminal.ClientFrame) error {
	switch input := frame.(type) {
	case terminal.TextFrame:
		return gateway.sessions.SendText(ctx, target, input.Text)
	case terminal.PasteFrame:
		return gateway.sessions.Paste(ctx, target, input.Text)
	case terminal.KeyFrame:
		return gateway.sessions.SendKey(ctx, target, input.Key, input.Modifiers)
	case terminal.ScrollFrame:
		return control.Scroll(input.Source, input.Direction, input.Lines, input.Column, input.Row)
	case terminal.ResizeFrame:
		return control.Resize(uint16(input.Columns), uint16(input.Rows))
	default:
		return terminal.ErrInvalidFrame
	}
}

func watchTerminalPing(ctx context.Context, connection *websocket.Conn, endReason *atomic.Uint32, cancelAttempt context.CancelFunc) {
	ticker := time.NewTicker(terminalPingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingContext, cancel := context.WithTimeout(ctx, terminalPongTimeout)
			err := connection.Ping(pingContext)
			cancel()
			if err != nil && ctx.Err() == nil {
				endReason.CompareAndSwap(0, 3)
				cancelAttempt()
				return
			}
		}
	}
}

func (gateway *Gateway) watchTerminalBearer(ctx context.Context, authorization string, endReason *atomic.Uint32, cancelAttempt context.CancelFunc) {
	ticker := time.NewTicker(terminalBearerInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			credential, err := gateway.bearer.Read()
			if (err != nil || !credential.Verify(authorization)) && ctx.Err() == nil {
				endReason.CompareAndSwap(0, 3)
				cancelAttempt()
				return
			}
		}
	}
}

func writeTerminalEnd(connection *websocket.Conn, code terminal.EndCode) {
	encoded, err := terminal.EncodeEnd(code)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = connection.Write(ctx, websocket.MessageText, encoded)
}

func (gateway *Gateway) registerLiveTerminal(id string, cancel context.CancelFunc) (*liveTerminal, bool) {
	gateway.liveMutex.Lock()
	defer gateway.liveMutex.Unlock()
	if gateway.closing {
		return nil, false
	}
	gateway.nextLiveTerminal++
	live := &liveTerminal{terminalID: id, cancel: cancel, done: make(chan struct{})}
	gateway.liveTerminals[gateway.nextLiveTerminal] = live
	return live, true
}

func (gateway *Gateway) unregisterLiveTerminal(target *liveTerminal) {
	gateway.liveMutex.Lock()
	for id, live := range gateway.liveTerminals {
		if live == target {
			delete(gateway.liveTerminals, id)
			break
		}
	}
	gateway.liveMutex.Unlock()
	close(target.done)
}

func (gateway *Gateway) closeLiveTerminals(id string) {
	gateway.liveMutex.Lock()
	for _, live := range gateway.liveTerminals {
		if live.terminalID == id {
			live.cancel()
		}
	}
	gateway.liveMutex.Unlock()
}

func (gateway *Gateway) CloseLiveTerminals(ctx context.Context) error {
	gateway.liveMutex.Lock()
	gateway.closing = true
	live := make([]*liveTerminal, 0, len(gateway.liveTerminals))
	for _, target := range gateway.liveTerminals {
		live = append(live, target)
		target.cancel()
	}
	gateway.liveMutex.Unlock()
	for _, target := range live {
		select {
		case <-target.done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
