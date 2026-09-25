// Package terminal owns the strict, bounded typed attachment wire.
package terminal

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/NielsdaWheelz/skidbladnir/internal/strictjson"
)

const (
	maximumFrameBytes       = 3 << 20
	MaximumClientFrameBytes = 2 << 20
	maximumANSIBytes        = 2 << 20
	maximumInputBytes       = 32768
	minimumColumns          = 20
	maximumColumns          = 1024
	minimumRows             = 5
	maximumRows             = 512
)

var (
	ErrInvalidFrame  = errors.New("invalid terminal frame")
	errFrameTooLarge = errors.New("terminal frame too large")
)

type ClientFrame interface{ isClientFrame() }
type TextFrame struct{ Text string }
type PasteFrame struct{ Text string }
type KeyFrame struct {
	Key       string
	Modifiers []string
}
type ScrollFrame struct {
	Source, Direction string
	Lines             uint16
	Column, Row       *uint16
}
type ResizeFrame struct {
	Columns, Rows             int
	CellWidthPx, CellHeightPx *uint32
}
type DetachFrame struct{}

func (TextFrame) isClientFrame()   {}
func (PasteFrame) isClientFrame()  {}
func (KeyFrame) isClientFrame()    {}
func (ScrollFrame) isClientFrame() {}
func (ResizeFrame) isClientFrame() {}
func (DetachFrame) isClientFrame() {}

type EndCode string

const (
	EndControlUnavailable EndCode = "control_unavailable"
	EndStreamLost         EndCode = "stream_lost"
	EndDetached           EndCode = "detached"
	EndProtocolError      EndCode = "protocol_error"
)

func required(encoded []byte, names ...string) bool {
	var fields map[string]json.RawMessage
	if strictjson.Decode(encoded, &fields) != nil {
		return false
	}
	for _, name := range names {
		value, ok := fields[name]
		if !ok || string(value) == "null" {
			return false
		}
	}
	for _, value := range fields {
		if string(value) == "null" {
			return false
		}
	}
	return true
}
func dimensions(columns, rows int) bool {
	return columns >= minimumColumns && columns <= maximumColumns && rows >= minimumRows && rows <= maximumRows
}
func endCode(code EndCode) bool {
	return code == EndControlUnavailable || code == EndStreamLost || code == EndDetached || code == EndProtocolError
}
func bounded(encoded []byte, limit int) error {
	if len(encoded) > limit {
		return errFrameTooLarge
	}
	return nil
}
func marshal(value any, limit int) ([]byte, error) {
	b, e := json.Marshal(value)
	if e != nil {
		return nil, ErrInvalidFrame
	}
	if e = bounded(b, limit); e != nil {
		return nil, e
	}
	return b, nil
}

func EncodeFrame(seq uint64, columns, rows int, full bool, ansi []byte) ([]byte, error) {
	if seq == 0 || !dimensions(columns, rows) || len(ansi) > maximumANSIBytes {
		return nil, ErrInvalidFrame
	}
	return marshal(struct {
		Kind       string `json:"kind"`
		Seq        string `json:"seq"`
		Columns    int    `json:"columns"`
		Rows       int    `json:"rows"`
		Full       bool   `json:"full"`
		ANSIBase64 string `json:"ansiBase64"`
	}{"Frame", strconv.FormatUint(seq, 10), columns, rows, full, base64.StdEncoding.EncodeToString(ansi)}, maximumFrameBytes)
}
func EncodeEnd(code EndCode) ([]byte, error) {
	if !endCode(code) {
		return nil, ErrInvalidFrame
	}
	return marshal(struct {
		Kind string  `json:"kind"`
		Code EndCode `json:"code"`
	}{"End", code}, maximumFrameBytes)
}
func ParseClientText(encoded []byte) (ClientFrame, error) {
	if err := bounded(encoded, MaximumClientFrameBytes); err != nil {
		return nil, err
	}
	var envelope struct {
		Kind string `json:"kind"`
	}
	if json.Unmarshal(encoded, &envelope) != nil {
		return nil, ErrInvalidFrame
	}
	switch envelope.Kind {
	case "Text", "Paste":
		var wire struct {
			Kind string `json:"kind"`
			Text string `json:"text"`
		}
		if strictjson.Decode(encoded, &wire) != nil || !required(encoded, "kind", "text") {
			return nil, ErrInvalidFrame
		}
		if wire.Kind == "Text" && validText(wire.Text) {
			return TextFrame{wire.Text}, nil
		}
		if wire.Kind == "Paste" && validPaste(wire.Text) {
			return PasteFrame{wire.Text}, nil
		}
	case "Key":
		var wire struct {
			Kind      string   `json:"kind"`
			Key       string   `json:"key"`
			Modifiers []string `json:"modifiers"`
		}
		if strictjson.Decode(encoded, &wire) == nil && required(encoded, "kind", "key", "modifiers") && wire.Modifiers != nil && validKey(wire.Key, wire.Modifiers) {
			return KeyFrame{wire.Key, wire.Modifiers}, nil
		}
	case "Scroll":
		var wire struct {
			Kind      string  `json:"kind"`
			Source    string  `json:"source"`
			Direction string  `json:"direction"`
			Lines     uint16  `json:"lines"`
			Column    *uint16 `json:"column,omitempty"`
			Row       *uint16 `json:"row,omitempty"`
		}
		if strictjson.Decode(encoded, &wire) == nil && required(encoded, "kind", "source", "direction", "lines") && validScroll(wire.Source, wire.Direction, wire.Lines, wire.Column, wire.Row) {
			return ScrollFrame{wire.Source, wire.Direction, wire.Lines, wire.Column, wire.Row}, nil
		}
	case "Resize":
		var wire struct {
			Kind         string  `json:"kind"`
			Columns      int     `json:"columns"`
			Rows         int     `json:"rows"`
			CellWidthPx  *uint32 `json:"cellWidthPx,omitempty"`
			CellHeightPx *uint32 `json:"cellHeightPx,omitempty"`
		}
		if strictjson.Decode(encoded, &wire) == nil && required(encoded, "kind", "columns", "rows") && dimensions(wire.Columns, wire.Rows) && (wire.CellWidthPx == nil || *wire.CellWidthPx > 0) && (wire.CellHeightPx == nil || *wire.CellHeightPx > 0) {
			return ResizeFrame{wire.Columns, wire.Rows, wire.CellWidthPx, wire.CellHeightPx}, nil
		}
	case "Detach":
		var wire struct {
			Kind string `json:"kind"`
		}
		if strictjson.Decode(encoded, &wire) == nil {
			return DetachFrame{}, nil
		}
	}
	return nil, ErrInvalidFrame
}
func validPaste(value string) bool {
	return value != "" && len(value) <= maximumInputBytes && utf8.ValidString(value)
}
func validText(value string) bool {
	if !validPaste(value) {
		return false
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}
func validKey(key string, modifiers []string) bool {
	switch key {
	case "enter", "escape", "tab", "backspace", "up", "down", "left", "right", "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12":
	default:
		r, n := utf8.DecodeRuneInString(key)
		if !utf8.ValidString(key) || n != len(key) || !unicode.IsPrint(r) || unicode.IsSpace(r) && r != ' ' {
			return false
		}
	}
	if len(modifiers) > 3 {
		return false
	}
	last := -1
	for _, m := range modifiers {
		index := -1
		switch m {
		case "ctrl":
			index = 0
		case "alt":
			index = 1
		case "shift":
			index = 2
		}
		if index <= last {
			return false
		}
		last = index
	}
	return true
}
func validScroll(source, direction string, lines uint16, column, row *uint16) bool {
	return (source == "wheel" || source == "page_key" && column == nil && row == nil) && (direction == "up" || direction == "down") && lines >= 1 && lines <= 512
}
