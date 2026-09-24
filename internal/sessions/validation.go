package sessions

import (
	"regexp"
	"slices"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

var (
	sessionNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)
)

func validateName(name string) error {
	if !sessionNamePattern.MatchString(name) {
		return newSessionError(ErrorNameInvalid, "Use 1–64 letters, numbers, underscores, or hyphens, beginning with a letter or number.")
	}
	return nil
}

func validateObjective(objective string) error {
	if objective == "" {
		return nil
	}
	if !utf8.ValidString(objective) || utf8.RuneCountInString(objective) > 240 || !norm.NFC.IsNormalString(objective) {
		return newSessionError(ErrorObjectiveInvalid, "Use 1–240 characters without terminal controls.")
	}
	for _, value := range objective {
		if isC0OrC1(value) || value == '\u2028' || value == '\u2029' || value >= '\u202a' && value <= '\u202e' || value >= '\u2066' && value <= '\u2069' {
			return newSessionError(ErrorObjectiveInvalid, "Use 1–240 characters without terminal controls.")
		}
	}
	return nil
}

func validWorkspaceLabel(value string) bool {
	if value == "" || !utf8.ValidString(value) || len(value) > 256 || utf8.RuneCountInString(value) > 64 ||
		normalizeWorkspaceNFC(value) != value || value[0] == ' ' || value[len(value)-1] == ' ' {
		return false
	}
	for _, symbol := range value {
		if symbol <= 0x1f || symbol >= 0x7f && symbol <= 0x9f || symbol == 0x061c ||
			symbol >= 0x200e && symbol <= 0x200f || symbol >= 0x2028 && symbol <= 0x202e ||
			symbol >= 0x2066 && symbol <= 0x2069 || symbol == 0xa0 || symbol == 0x1680 ||
			symbol >= 0x2000 && symbol <= 0x200a || symbol == 0x202f || symbol == 0x205f || symbol == 0x3000 {
			return false
		}
	}
	return true
}

// x/text's whole-string NFC inserts CGJ after 30 nonstarters. This checks
// canonical composition without rewriting otherwise valid workspace labels.
func normalizeWorkspaceNFC(value string) string {
	type scalar struct {
		value rune
		class uint8
	}
	decomposed := make([]scalar, 0, len(value))
	for _, symbol := range value {
		for _, part := range norm.NFD.String(string(symbol)) {
			decomposed = append(decomposed, scalar{part, norm.NFD.PropertiesString(string(part)).CCC()})
		}
	}
	for start := 0; start < len(decomposed); {
		if decomposed[start].class == 0 {
			start++
			continue
		}
		end := start + 1
		for end < len(decomposed) && decomposed[end].class != 0 {
			end++
		}
		slices.SortStableFunc(decomposed[start:end], func(left, right scalar) int {
			return int(left.class) - int(right.class)
		})
		start = end
	}
	composed := make([]rune, 0, len(decomposed))
	starter := -1
	var blockingClass uint8
	for _, part := range decomposed {
		if starter >= 0 && (blockingClass == 0 || blockingClass < part.class) {
			pair := norm.NFC.String(string(composed[starter]) + string(part.value))
			combined, size := utf8.DecodeRuneInString(pair)
			if size == len(pair) {
				composed[starter] = combined
				continue
			}
		}
		if part.class == 0 {
			starter = len(composed)
		}
		composed = append(composed, part.value)
		blockingClass = part.class
	}
	return string(composed)
}

func isC0OrC1(value rune) bool {
	return value >= 0 && value <= 0x1f || value >= 0x7f && value <= 0x9f
}

func newSessionError(code ErrorCode, message string) *Error {
	return &Error{Code: code, Message: message, Dispatch: "not_sent"}
}
