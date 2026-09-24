package profile

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

var environmentPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

const maxKeyLength = 32

type Provider string

const (
	ProviderCodex  Provider = "Codex"
	ProviderClaude Provider = "Claude"
)

func ParseProvider(value string) (Provider, error) {
	provider := Provider(value)
	switch provider {
	case ProviderCodex, ProviderClaude:
		return provider, nil
	default:
		return "", errors.New("provider is invalid")
	}
}

type Key string

func validKey(value string) bool {
	if len(value) == 0 || len(value) > maxKeyLength {
		return false
	}
	for index := range len(value) {
		character := value[index]
		if character < 'a' || character > 'z' {
			if index == 0 || !(character >= '0' && character <= '9' || character == '_' || character == '-') {
				return false
			}
		}
	}
	return true
}

type EnvironmentVariable struct {
	Name  string
	Value string
}

// Profile names one account for an agent that herdr starts in a new pane. The
// pane is created with Environment, which carries the provider's account home;
// the pane shell's own command resolution and aliases decide what runs.
type Profile struct {
	Key         Key
	Label       string
	Provider    Provider
	Environment []EnvironmentVariable
}

func Validate(profiles []Profile) ([]Profile, error) {
	validated := make([]Profile, 0, len(profiles))
	keys := make(map[Key]struct{}, len(profiles))
	labels := make(map[string]struct{}, len(profiles))
	homes := map[Provider]map[string]struct{}{
		ProviderCodex:  {},
		ProviderClaude: {},
	}
	for _, profile := range profiles {
		if !validKey(string(profile.Key)) {
			return nil, errors.New("profile key is invalid")
		}
		if _, found := keys[profile.Key]; found {
			return nil, errors.New("profile key is duplicated")
		}
		if !safeLabel(profile.Label) {
			return nil, fmt.Errorf("profile %s label is invalid", profile.Key)
		}
		if _, found := labels[profile.Label]; found {
			return nil, errors.New("profile label is duplicated")
		}
		if _, err := ParseProvider(string(profile.Provider)); err != nil {
			return nil, fmt.Errorf("profile %s provider is invalid", profile.Key)
		}
		homeName := "CODEX_HOME"
		if profile.Provider == ProviderClaude {
			homeName = "CLAUDE_CONFIG_DIR"
		}
		home := ""
		environmentNames := make(map[string]struct{}, len(profile.Environment))
		for _, variable := range profile.Environment {
			if !environmentPattern.MatchString(variable.Name) || !utf8.ValidString(variable.Value) || strings.ContainsRune(variable.Value, 0) {
				return nil, fmt.Errorf("profile %s environment is invalid", profile.Key)
			}
			if _, found := environmentNames[variable.Name]; found {
				return nil, fmt.Errorf("profile %s environment name is duplicated", profile.Key)
			}
			environmentNames[variable.Name] = struct{}{}
			if variable.Name == "CODEX_HOME" || variable.Name == "CLAUDE_CONFIG_DIR" {
				if variable.Name != homeName {
					return nil, fmt.Errorf("profile %s declares another provider's home", profile.Key)
				}
				home = variable.Value
			}
		}
		if home == "" || !filepath.IsAbs(home) {
			return nil, fmt.Errorf("profile %s provider home must be absolute", profile.Key)
		}
		if _, found := homes[profile.Provider][home]; found {
			return nil, fmt.Errorf("provider %s home is duplicated", profile.Provider)
		}
		profile.Environment = append([]EnvironmentVariable(nil), profile.Environment...)
		validated = append(validated, profile)
		keys[profile.Key] = struct{}{}
		labels[profile.Label] = struct{}{}
		homes[profile.Provider][home] = struct{}{}
	}
	return validated, nil
}

func Clone(profiles []Profile) []Profile {
	cloned := make([]Profile, len(profiles))
	for index, profile := range profiles {
		profile.Environment = append([]EnvironmentVariable(nil), profile.Environment...)
		cloned[index] = profile
	}
	return cloned
}

func safeLabel(label string) bool {
	if strings.TrimSpace(label) == "" || !utf8.ValidString(label) || utf8.RuneCountInString(label) > 64 || !norm.NFC.IsNormalString(label) {
		return false
	}
	for _, value := range label {
		if unicode.IsControl(value) || unicode.Is(unicode.Bidi_Control, value) || value == ' ' || value == ' ' {
			return false
		}
	}
	return true
}
