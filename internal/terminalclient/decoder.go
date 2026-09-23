package terminalclient

import (
	"bytes"
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/NielsdaWheelz/skidbladnir/internal/terminal"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

const maxSequenceBytes = 4096

var pasteStart = []byte("\x1b[200~")
var pasteEnd = []byte("\x1b[201~")
var errInput = errors.New("undecodable terminal input; attachment ended")

type decodedInput struct {
	frames []terminal.ClientFrame
	notice string
	detach bool
}
type inputDecoder struct {
	pending []byte
	paste   []byte
	inPaste bool
	prefix  bool
	decoder uv.EventDecoder
	rows    int
}

func (d *inputDecoder) Feed(data []byte) (decodedInput, error) {
	var result decodedInput
	for _, b := range data {
		if d.inPaste {
			d.paste = append(d.paste, b)
			suffix := 0
			for n := 1; n < len(pasteEnd) && n <= len(d.paste); n++ {
				if bytes.Equal(d.paste[len(d.paste)-n:], pasteEnd[:n]) {
					suffix = n
				}
			}
			if len(d.paste)-suffix > terminal.MaximumInputBytes {
				return result, errors.New("paste exceeds 32 kib; attachment ended")
			}
			if bytes.HasSuffix(d.paste, pasteEnd) {
				content := d.paste[:len(d.paste)-len(pasteEnd)]
				if !utf8.Valid(content) {
					return result, errInput
				}
				if len(content) > 0 {
					result.frames = append(result.frames, terminal.PasteFrame{Text: string(content)})
				}
				d.paste = nil
				d.inPaste = false
			}
			continue
		}
		d.pending = append(d.pending, b)
		if len(d.pending) > maxSequenceBytes {
			return result, errInput
		}
		for len(d.pending) > 0 {
			if bytes.Equal(d.pending, pasteStart) {
				d.pending = nil
				if d.prefix {
					result.frames = append(result.frames, terminal.KeyFrame{Key: "]", Modifiers: []string{"ctrl"}})
					d.prefix = false
				}
				d.inPaste = true
				break
			}
			if len(d.pending) < len(pasteStart) && bytes.Equal(d.pending, pasteStart[:len(d.pending)]) {
				break
			}
			if d.pending[0] >= utf8.RuneSelf && !utf8.FullRune(d.pending) {
				break
			}
			sequence, _, n, state := ansi.DecodeSequence(d.pending, 0, nil)
			if n == 0 || state != byte(ansi.NormalState) {
				break
			}
			if n > len(d.pending) || len(sequence) == 0 {
				return result, errInput
			}
			d.pending = d.pending[n:]
			event, err := d.event(sequence)
			if err != nil {
				return result, err
			}
			if event.notice != "" {
				result.notice = event.notice
			}
			if event.detach {
				result.detach = true
				return result, nil
			}
			result.frames = append(result.frames, event.frames...)
		}
	}
	return result, nil
}
func (d *inputDecoder) FlushEscape() (decodedInput, error) {
	if bytes.Equal(d.pending, []byte{0x1b}) {
		d.pending = nil
		return d.key(uv.KeyPressEvent{Code: uv.KeyEscape})
	}
	if len(d.pending) > 0 {
		return decodedInput{}, errInput
	}
	return decodedInput{}, nil
}
func (d *inputDecoder) event(sequence []byte) (decodedInput, error) {
	if sequence[0] >= utf8.RuneSelf {
		if !utf8.Valid(sequence) {
			return decodedInput{}, errInput
		}
		frames := []terminal.ClientFrame{terminal.TextFrame{Text: string(sequence)}}
		if d.prefix {
			d.prefix = false
			frames = append([]terminal.ClientFrame{terminal.KeyFrame{Key: "]", Modifiers: []string{"ctrl"}}}, frames...)
		}
		return decodedInput{frames: frames}, nil
	}
	n, event := d.decoder.Decode(sequence)
	if n != len(sequence) || event == nil {
		return decodedInput{}, errInput
	}
	switch value := event.(type) {
	case uv.KeyPressEvent:
		return d.key(value)
	case uv.KeyReleaseEvent:
		return decodedInput{}, nil
	case uv.CursorPositionEvent, uv.PrimaryDeviceAttributesEvent, uv.SecondaryDeviceAttributesEvent, uv.TertiaryDeviceAttributesEvent, uv.ModeReportEvent, uv.WindowOpEvent, uv.ForegroundColorEvent, uv.BackgroundColorEvent, uv.CursorColorEvent, uv.TerminalVersionEvent, uv.KeyboardEnhancementsEvent, uv.CapabilityEvent, uv.FocusEvent, uv.BlurEvent, uv.PixelSizeEvent, uv.CellSizeEvent:
		return decodedInput{}, nil
	case uv.MultiEvent:
		return decodedInput{}, errors.New("ambiguous terminal key or reply; attachment ended")
	default:
		return decodedInput{}, errInput
	}
}
func (d *inputDecoder) key(event uv.KeyPressEvent) (decodedInput, error) {
	if d.prefix {
		d.prefix = false
		if event.Mod == 0 && (event.Text == "d" || event.Code == 'd') {
			return decodedInput{detach: true}, nil
		}
		next, err := d.key(event)
		if err != nil {
			return next, err
		}
		next.frames = append([]terminal.ClientFrame{terminal.KeyFrame{Key: "]", Modifiers: []string{"ctrl"}}}, next.frames...)
		return next, nil
	}
	mod := event.Mod
	if mod & ^(uv.ModCtrl|uv.ModAlt|uv.ModShift|uv.ModCapsLock|uv.ModNumLock) != 0 {
		return decodedInput{notice: "unsupported keyboard modifier"}, nil
	}
	mods := make([]string, 0, 3)
	if mod.Contains(uv.ModCtrl) {
		mods = append(mods, "ctrl")
	}
	if mod.Contains(uv.ModAlt) {
		mods = append(mods, "alt")
	}
	if mod.Contains(uv.ModShift) {
		mods = append(mods, "shift")
	}
	key := ""
	switch event.Code {
	case uv.KeyEnter:
		key = "enter"
	case uv.KeyEscape:
		key = "escape"
	case uv.KeyTab:
		key = "tab"
	case uv.KeyBackspace:
		key = "backspace"
	case uv.KeyUp:
		key = "up"
	case uv.KeyDown:
		key = "down"
	case uv.KeyLeft:
		key = "left"
	case uv.KeyRight:
		key = "right"
	case uv.KeyPgUp, uv.KeyPgDown:
		if mod&(uv.ModCtrl|uv.ModAlt|uv.ModShift) != 0 {
			return decodedInput{notice: "modified page keys are unavailable"}, nil
		}
		direction := "up"
		if event.Code == uv.KeyPgDown {
			direction = "down"
		}
		return decodedInput{frames: []terminal.ClientFrame{terminal.ScrollFrame{Source: "page_key", Direction: direction, Lines: uint16(d.rows)}}}, nil
	case uv.KeyHome, uv.KeyEnd, uv.KeyInsert, uv.KeyDelete:
		return decodedInput{notice: "home, end, insert and forward delete are unavailable in this herdr version"}, nil
	default:
		if event.Code >= uv.KeyF1 && event.Code <= uv.KeyF12 {
			key = fmt.Sprintf("f%d", event.Code-uv.KeyF1+1)
		} else if event.Text != "" {
			if !mod.Contains(uv.ModCtrl) && !mod.Contains(uv.ModAlt) {
				return decodedInput{frames: []terminal.ClientFrame{terminal.TextFrame{Text: event.Text}}}, nil
			}
			r, n := utf8.DecodeRuneInString(event.Text)
			if !utf8.ValidString(event.Text) || n != len(event.Text) {
				return decodedInput{}, errInput
			}
			if unicode.IsSpace(r) && r != ' ' {
				return decodedInput{notice: "modified unicode whitespace key is unavailable in this herdr version"}, nil
			}
			key = event.Text
		} else if unicode.IsPrint(event.Code) {
			if unicode.IsSpace(event.Code) && event.Code != ' ' {
				return decodedInput{notice: "modified unicode whitespace key is unavailable in this herdr version"}, nil
			}
			key = string(event.Code)
		} else {
			return decodedInput{notice: "unsupported keyboard key"}, nil
		}
	}
	if key == "]" && mod == uv.ModCtrl {
		d.prefix = true
		return decodedInput{}, nil
	}
	return decodedInput{frames: []terminal.ClientFrame{terminal.KeyFrame{Key: key, Modifiers: mods}}}, nil
}
