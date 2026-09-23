package agentruntime

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	processinfo "github.com/NielsdaWheelz/skidbladnir/internal/process"
	"golang.org/x/text/unicode/norm"
)

const (
	RegistrationToken      = "skid_agent_runtime"
	registrationValueLimit = 80

	maxPIDTextLength                  = 19
	maxStartIdentityTextLength        = 20
	maxProviderTextLength             = len("Claude")
	maxProviderSessionIDLength        = 128
	maxEncodedProviderSessionIDLength = (maxProviderSessionIDLength*8 + 5) / 6
	maxEncodedRegistrationLength      = len("v1") + 5 + maxPIDTextLength + maxStartIdentityTextLength + maxProviderTextLength + maxProfileKeyTextLength + maxEncodedProviderSessionIDLength
)

var registrationFragments = [...]string{
	"skid_agent_runtime_00",
	"skid_agent_runtime_01",
	"skid_agent_runtime_02",
	"skid_agent_runtime_03",
}

type RegistrationFacts struct {
	Profile         ProfileKey
	ProviderSession *ProviderSessionFacts
}

// ProjectRegistration accepts only a registration for the currently observed
// foreground process. A launch row or provider label alone never proves it.
func ProjectRegistration(profiles []Profile, observation processinfo.Observation, provider Provider, tokens map[string]string) (RegistrationFacts, bool) {
	encoded, complete := readRegistration(tokens)
	if !complete {
		return RegistrationFacts{}, false
	}
	foreground := Foreground{Provider: provider, PID: observation.PID, StartIdentity: observation.StartIdentity}
	registration, valid := acceptRegistration(profiles, foreground, encoded)
	if !valid {
		return RegistrationFacts{}, false
	}
	facts := RegistrationFacts{Profile: registration.profile}
	if provider == ProviderClaude {
		name := claudeName(observation.Argv)
		if registration.providerSessionID != "" || name != "" {
			facts.ProviderSession, _ = NewProviderSessionFacts(registration.providerSessionID, name)
		}
	}
	return facts, true
}

type Foreground struct {
	Provider      Provider
	PID           processinfo.PID
	StartIdentity processinfo.StartIdentity
}

type ProviderSessionFacts struct {
	id   string
	name string
}

func NewProviderSessionFacts(id, name string) (*ProviderSessionFacts, error) {
	if err := validateProviderSessionFacts(id, name); err != nil {
		return nil, err
	}
	return &ProviderSessionFacts{id: id, name: name}, nil
}

func validateProviderSessionFacts(id, name string) error {
	if id == "" && name == "" {
		return errors.New("provider session facts are empty")
	}
	if id != "" && !validProviderSessionID(id) {
		return errors.New("provider session id is invalid")
	}
	if name != "" && !validProviderSessionName(name) {
		return errors.New("provider session name is invalid")
	}
	return nil
}

func (facts ProviderSessionFacts) ID() string   { return facts.id }
func (facts ProviderSessionFacts) Name() string { return facts.name }

func ClassifyForeground(profiles []Profile, observation processinfo.Observation) (Foreground, bool) {
	if observation.PID <= 0 || !validStartIdentity(observation.StartIdentity) {
		return Foreground{}, false
	}
	var provider Provider
	for _, profile := range profiles {
		for _, signature := range profile.ForegroundSignatures {
			if !matchesSignature(observation, signature) {
				continue
			}
			if provider != "" && provider != profile.Provider {
				return Foreground{}, false
			}
			provider = profile.Provider
		}
	}
	if provider == "" {
		return Foreground{}, false
	}
	return Foreground{Provider: provider, PID: observation.PID, StartIdentity: observation.StartIdentity}, true
}

func HookOrigin(
	profiles []Profile,
	ancestry []processinfo.Observation,
	foregroundPGID processinfo.PID,
	foregroundPIDs []processinfo.PID,
) (Foreground, bool) {
	if foregroundPGID <= 0 || len(foregroundPIDs) == 0 {
		return Foreground{}, false
	}
	type match struct {
		foreground  Foreground
		observation processinfo.Observation
	}
	matches := make([]match, 0, 2)
	for _, observation := range ancestry {
		listed := false
		for _, pid := range foregroundPIDs {
			listed = listed || pid == observation.PID
		}
		if !listed || observation.ProcessGroup != foregroundPGID || observation.ForegroundProcessGroup != foregroundPGID {
			continue
		}
		if foreground, found := ClassifyForeground(profiles, observation); found {
			matches = append(matches, match{foreground: foreground, observation: observation})
		}
	}

	var origin match
	switch len(matches) {
	case 1:
		origin = matches[0]
	case 2:
		native, wrapper := matches[0], matches[1]
		if native.foreground.Provider != ProviderCodex || wrapper.foreground.Provider != ProviderCodex ||
			native.observation.ExecutableBase() != "codex" || wrapper.observation.ExecutableBase() != "node" ||
			native.observation.ParentPID != wrapper.observation.PID {
			return Foreground{}, false
		}
		origin = wrapper
	default:
		return Foreground{}, false
	}
	return origin.foreground, true
}

func EncodeRegistration(foreground Foreground, profile ProfileKey, providerSessionID string) (map[string]*string, error) {
	if foreground.PID <= 0 || !validStartIdentity(foreground.StartIdentity) {
		return nil, errors.New("foreground lifetime is invalid")
	}
	if _, err := ParseProvider(foreground.Provider.String()); err != nil {
		return nil, errors.New("foreground provider is invalid")
	}
	profileValue := "-"
	if profile != "" {
		if _, err := ParseProfileKey(string(profile)); err != nil {
			return nil, errors.New("runtime profile is invalid")
		}
		profileValue = string(profile)
	}
	if !validProviderSessionID(providerSessionID) {
		return nil, errors.New("provider session id is invalid")
	}
	encoded := strings.Join([]string{
		"v1",
		strconv.Itoa(int(foreground.PID)),
		string(foreground.StartIdentity),
		foreground.Provider.String(),
		profileValue,
		base64.RawURLEncoding.EncodeToString([]byte(providerSessionID)),
	}, ":")
	if len(encoded) > len(registrationFragments)*registrationValueLimit {
		return nil, errors.New("agent registration exceeds herdr metadata")
	}
	count := (len(encoded) + registrationValueLimit - 1) / registrationValueLimit
	digest := sha256.Sum256([]byte(encoded))
	header := "v2:" + strconv.Itoa(count) + ":" + hex.EncodeToString(digest[:])
	tokens := map[string]*string{RegistrationToken: &header}
	for index, key := range registrationFragments {
		if index >= count {
			tokens[key] = nil
			continue
		}
		start := index * registrationValueLimit
		end := min(start+registrationValueLimit, len(encoded))
		part := encoded[start:end]
		tokens[key] = &part
	}
	return tokens, nil
}

func readRegistration(tokens map[string]string) (string, bool) {
	header := strings.Split(tokens[RegistrationToken], ":")
	if len(header) != 3 || header[0] != "v2" {
		return "", false
	}
	count, err := strconv.Atoi(header[1])
	if err != nil || count < 1 || count > len(registrationFragments) || strconv.Itoa(count) != header[1] {
		return "", false
	}
	var encoded strings.Builder
	for index, key := range registrationFragments {
		part, present := tokens[key]
		if index >= count {
			if present {
				return "", false
			}
			continue
		}
		if !present || len(part) == 0 || len(part) > registrationValueLimit || index < count-1 && len(part) != registrationValueLimit {
			return "", false
		}
		encoded.WriteString(part)
	}
	if encoded.Len() > maxEncodedRegistrationLength {
		return "", false
	}
	digest := sha256.Sum256([]byte(encoded.String()))
	if header[2] != hex.EncodeToString(digest[:]) {
		return "", false
	}
	return encoded.String(), true
}

type registration struct {
	pid               processinfo.PID
	startIdentity     processinfo.StartIdentity
	provider          Provider
	profile           ProfileKey
	providerSessionID string
}

func acceptRegistration(profiles []Profile, foreground Foreground, encoded string) (registration, bool) {
	parsed, valid := parseRegistration(encoded)
	if !valid || parsed.pid != foreground.PID || parsed.startIdentity != foreground.StartIdentity || parsed.provider != foreground.Provider {
		return registration{}, false
	}
	if parsed.profile != "" {
		if !configuredProfileMatchesProvider(profiles, parsed.profile, parsed.provider) {
			return registration{}, false
		}
	}
	return parsed, true
}

func parseRegistration(encoded string) (registration, bool) {
	if len(encoded) > maxEncodedRegistrationLength {
		return registration{}, false
	}
	fields := strings.Split(encoded, ":")
	if len(fields) != 6 || fields[0] != "v1" {
		return registration{}, false
	}
	pid, err := strconv.Atoi(fields[1])
	if err != nil || pid <= 0 || strconv.Itoa(pid) != fields[1] || !validStartIdentity(processinfo.StartIdentity(fields[2])) {
		return registration{}, false
	}
	provider, err := ParseProvider(fields[3])
	if err != nil {
		return registration{}, false
	}
	var profile ProfileKey
	if fields[4] != "-" {
		parsedProfile, err := ParseProfileKey(fields[4])
		if err != nil {
			return registration{}, false
		}
		profile = parsedProfile
	}
	if len(fields[5]) > maxEncodedProviderSessionIDLength {
		return registration{}, false
	}
	providerSessionID, err := base64.RawURLEncoding.DecodeString(fields[5])
	if err != nil || base64.RawURLEncoding.EncodeToString(providerSessionID) != fields[5] || !validProviderSessionID(string(providerSessionID)) {
		return registration{}, false
	}
	return registration{
		pid:               processinfo.PID(pid),
		startIdentity:     processinfo.StartIdentity(fields[2]),
		provider:          provider,
		profile:           profile,
		providerSessionID: string(providerSessionID),
	}, true
}

func configuredProfileMatchesProvider(profiles []Profile, key ProfileKey, provider Provider) bool {
	for _, profile := range profiles {
		if profile.Key == key && profile.Provider == provider {
			return true
		}
	}
	return false
}

func validStartIdentity(value processinfo.StartIdentity) bool {
	parsed, err := strconv.ParseUint(string(value), 10, 64)
	return err == nil && parsed > 0 && strconv.FormatUint(parsed, 10) == string(value)
}

func matchesSignature(observation processinfo.Observation, signature ForegroundSignature) bool {
	return (signature.ExecutableBase == "" || observation.ExecutableBase() == signature.ExecutableBase) &&
		(signature.Argument0 == "" || observation.Argument(0) == signature.Argument0) &&
		(signature.Argument1 == "" || observation.Argument(1) == signature.Argument1)
}

func claudeName(argv []string) string {
	name := ""
	found := false
	for index := 1; index < len(argv); index++ {
		argument := argv[index]
		if argument == "--" {
			break
		}
		candidate := ""
		switch {
		case argument == "-n" || argument == "--name":
			if index+1 >= len(argv) {
				return ""
			}
			index++
			candidate = argv[index]
		case strings.HasPrefix(argument, "--name="):
			candidate = strings.TrimPrefix(argument, "--name=")
		default:
			continue
		}
		if found || !validProviderSessionName(candidate) {
			return ""
		}
		name, found = candidate, true
	}
	return name
}

func validProviderSessionID(value string) bool {
	if len(value) == 0 || len(value) > maxProviderSessionIDLength {
		return false
	}
	for index := range len(value) {
		if value[index] < 0x21 || value[index] > 0x7e {
			return false
		}
	}
	return true
}

func validProviderSessionName(value string) bool {
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) == 0 || utf8.RuneCountInString(value) > 128 || !norm.NFC.IsNormalString(value) {
		return false
	}
	for _, character := range value {
		if forbiddenIdentityTextRune(character) {
			return false
		}
	}
	return true
}

func forbiddenIdentityTextRune(character rune) bool {
	return unicode.IsControl(character) || unicode.Is(unicode.Bidi_Control, character) ||
		character == '\u2028' || character == '\u2029'
}
