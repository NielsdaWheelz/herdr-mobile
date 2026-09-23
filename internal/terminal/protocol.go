// Package terminal owns the strict, bounded typed attachment wire.
package terminal

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/NielsdaWheelz/skidbladnir/internal/strictjson"
)

const (
	MaximumFrameBytes       = 3 << 20
	MaximumClientFrameBytes = 2 << 20
	MaximumANSIBytes        = 2 << 20
	MaximumInputBytes       = 32768
	MinimumColumns          = 20
	MaximumColumns          = 1024
	MinimumRows             = 5
	MaximumRows             = 512
)

var (
	ErrInvalidFrame  = errors.New("invalid terminal frame")
	ErrFrameTooLarge = errors.New("terminal frame too large")
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

type ServerFrame interface{ isServerFrame() }
type Frame struct {
	Seq           uint64
	Columns, Rows int
	Full          bool
	ANSI          []byte
}
type EndCode string

const (
	EndControlUnavailable EndCode = "control_unavailable"
	EndStreamLost         EndCode = "stream_lost"
	EndDetached           EndCode = "detached"
	EndProtocolError      EndCode = "protocol_error"
)

type EndFrame struct{ Code EndCode }

func (Frame) isServerFrame()    {}
func (EndFrame) isServerFrame() {}

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
	return columns >= MinimumColumns && columns <= MaximumColumns && rows >= MinimumRows && rows <= MaximumRows
}
func endCode(code EndCode) bool {
	return code == EndControlUnavailable || code == EndStreamLost || code == EndDetached || code == EndProtocolError
}
func bounded(encoded []byte, limit int) error {
	if len(encoded) > limit {
		return ErrFrameTooLarge
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
	if seq == 0 || !dimensions(columns, rows) || len(ansi) > MaximumANSIBytes {
		return nil, ErrInvalidFrame
	}
	return marshal(struct {
		Kind       string `json:"kind"`
		Seq        string `json:"seq"`
		Columns    int    `json:"columns"`
		Rows       int    `json:"rows"`
		Full       bool   `json:"full"`
		ANSIBase64 string `json:"ansiBase64"`
	}{"Frame", strconv.FormatUint(seq, 10), columns, rows, full, base64.StdEncoding.EncodeToString(ansi)}, MaximumFrameBytes)
}
func EncodeEnd(code EndCode) ([]byte, error) {
	if !endCode(code) {
		return nil, ErrInvalidFrame
	}
	return marshal(struct {
		Kind string  `json:"kind"`
		Code EndCode `json:"code"`
	}{"End", code}, MaximumFrameBytes)
}
func EncodeResize(columns, rows int) ([]byte, error) {
	if !dimensions(columns, rows) {
		return nil, ErrInvalidFrame
	}
	return marshal(struct {
		Kind    string `json:"kind"`
		Columns int    `json:"columns"`
		Rows    int    `json:"rows"`
	}{"Resize", columns, rows}, MaximumClientFrameBytes)
}
func EncodeDetach() []byte { return []byte(`{"kind":"Detach"}`) }
func EncodeText(value string) ([]byte, error) {
	if !validText(value) {
		return nil, ErrInvalidFrame
	}
	return marshal(struct {
		Kind string `json:"kind"`
		Text string `json:"text"`
	}{"Text", value}, MaximumClientFrameBytes)
}
func EncodePaste(value string) ([]byte, error) {
	if !validPaste(value) {
		return nil, ErrInvalidFrame
	}
	return marshal(struct {
		Kind string `json:"kind"`
		Text string `json:"text"`
	}{"Paste", value}, MaximumClientFrameBytes)
}
func EncodeKey(key string, modifiers []string) ([]byte, error) {
	if modifiers == nil {
		modifiers = []string{}
	}
	if !validKey(key, modifiers) {
		return nil, ErrInvalidFrame
	}
	return marshal(struct {
		Kind      string   `json:"kind"`
		Key       string   `json:"key"`
		Modifiers []string `json:"modifiers"`
	}{"Key", key, modifiers}, MaximumClientFrameBytes)
}
func EncodeScroll(source, direction string, lines uint16, column, row *uint16) ([]byte, error) {
	if !validScroll(source, direction, lines, column, row) {
		return nil, ErrInvalidFrame
	}
	return marshal(struct {
		Kind      string  `json:"kind"`
		Source    string  `json:"source"`
		Direction string  `json:"direction"`
		Lines     uint16  `json:"lines"`
		Column    *uint16 `json:"column,omitempty"`
		Row       *uint16 `json:"row,omitempty"`
	}{"Scroll", source, direction, lines, column, row}, MaximumClientFrameBytes)
}

func ParseServerText(encoded []byte) (ServerFrame, error) {
	if err := bounded(encoded, MaximumFrameBytes); err != nil {
		return nil, err
	}
	var envelope struct {
		Kind string `json:"kind"`
	}
	if json.Unmarshal(encoded, &envelope) != nil {
		return nil, ErrInvalidFrame
	}
	switch envelope.Kind {
	case "Frame":
		var wire struct {
			Kind       string `json:"kind"`
			Seq        string `json:"seq"`
			Columns    int    `json:"columns"`
			Rows       int    `json:"rows"`
			Full       bool   `json:"full"`
			ANSIBase64 string `json:"ansiBase64"`
		}
		if strictjson.Decode(encoded, &wire) != nil || !required(encoded, "kind", "seq", "columns", "rows", "full", "ansiBase64") || !dimensions(wire.Columns, wire.Rows) || wire.Seq == "" || wire.Seq[0] == '0' || strings.Trim(wire.Seq, "0123456789") != "" {
			return nil, ErrInvalidFrame
		}
		seq, err := strconv.ParseUint(wire.Seq, 10, 64)
		if err != nil || seq == 0 {
			return nil, ErrInvalidFrame
		}
		data, err := base64.StdEncoding.Strict().DecodeString(wire.ANSIBase64)
		if err != nil || len(data) > MaximumANSIBytes || base64.StdEncoding.EncodeToString(data) != wire.ANSIBase64 {
			return nil, ErrInvalidFrame
		}
		return Frame{seq, wire.Columns, wire.Rows, wire.Full, data}, nil
	case "End":
		var wire struct {
			Kind string  `json:"kind"`
			Code EndCode `json:"code"`
		}
		if strictjson.Decode(encoded, &wire) != nil || !required(encoded, "kind", "code") || !endCode(wire.Code) {
			return nil, ErrInvalidFrame
		}
		return EndFrame{wire.Code}, nil
	default:
		return nil, ErrInvalidFrame
	}
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
	return value != "" && len(value) <= MaximumInputBytes && utf8.ValidString(value)
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
