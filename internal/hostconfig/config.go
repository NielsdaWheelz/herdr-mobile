package hostconfig

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/NielsdaWheelz/skidbladnir/internal/platform"
	"github.com/NielsdaWheelz/skidbladnir/internal/profile"
	"github.com/NielsdaWheelz/skidbladnir/internal/strictjson"
)

const maximumConfigBytes = 64 * 1024

var expectedProfiles = [...]struct {
	key      string
	provider profile.Provider
}{
	{key: "personal", provider: profile.ProviderCodex},
	{key: "work", provider: profile.ProviderCodex},
	{key: "work2", provider: profile.ProviderCodex},
	{key: "claude-work", provider: profile.ProviderClaude},
}

type Config struct {
	Herdr    HerdrConfig
	Profiles []profile.Profile
}

type HerdrConfig struct {
	Path          string
	SocketPath    string
	TestedVersion string
}

func Load(path string, runtime platform.Kind) (config Config, resultErr error) {
	if path == "" {
		return Config{}, errors.New("host config path is empty")
	}
	// Host configuration is a deployment-owned local regular file. Nonblocking,
	// no-follow admission rejects FIFOs, devices, and symlinks before startup can
	// wait on an unbounded filesystem producer; the capped read prevents a large
	// file from being normalized into memory before its size is rejected.
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return Config{}, fmt.Errorf("read host config: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			config = Config{}
			resultErr = errors.Join(resultErr, fmt.Errorf("close host config: %w", closeErr))
		}
	}()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return Config{}, errors.New("host config is not a regular file")
	}
	encoded, err := io.ReadAll(io.LimitReader(file, maximumConfigBytes+1))
	if err != nil {
		return Config{}, fmt.Errorf("read host config: %w", err)
	}
	if len(encoded) > maximumConfigBytes {
		return Config{}, errors.New("host config is too large")
	}
	config, err = parse(encoded, runtime)
	if err != nil {
		return Config{}, fmt.Errorf("parse host config: %w", err)
	}
	return config, nil
}

func parse(encoded []byte, runtime platform.Kind) (Config, error) {
	if len(encoded) == 0 || len(encoded) > maximumConfigBytes {
		return Config{}, errors.New("host config has invalid size")
	}
	if runtime != platform.KindLinux && runtime != platform.KindDarwin {
		return Config{}, errors.New("runtime platform is unsupported")
	}
	var wire *configDTO
	if err := strictjson.Decode(encoded, &wire); err != nil || wire == nil {
		return Config{}, errors.New("host config is not canonical JSON")
	}
	return wire.validate(runtime)
}

type configDTO struct {
	Platform stringField   `json:"platform"`
	Herdr    *herdrDTO     `json:"herdr"`
	Profiles *[]profileDTO `json:"profiles"`
}

type herdrDTO struct {
	Path          stringField `json:"path"`
	SocketPath    stringField `json:"socketPath"`
	TestedVersion stringField `json:"testedVersion"`
}

type profileDTO struct {
	Key         stringField               `json:"key"`
	Label       stringField               `json:"label"`
	Provider    stringField               `json:"provider"`
	Environment *[]environmentVariableDTO `json:"environment"`
}

type environmentVariableDTO struct {
	Name  stringField `json:"name"`
	Value stringField `json:"value"`
}

func (wire configDTO) validate(runtime platform.Kind) (Config, error) {
	if !wire.Platform.present || wire.Herdr == nil || wire.Profiles == nil {
		return Config{}, errors.New("host config omits a required member")
	}
	kind := platform.Kind(wire.Platform.value)
	if kind != platform.KindLinux && kind != platform.KindDarwin {
		return Config{}, errors.New("host config platform is unsupported")
	}
	if kind != runtime {
		return Config{}, fmt.Errorf("host config platform %q does not match runtime %q", kind, runtime)
	}
	if !wire.Herdr.Path.present || !wire.Herdr.SocketPath.present || !wire.Herdr.TestedVersion.present ||
		!validAbsolutePath(wire.Herdr.Path.value) || !validAbsolutePath(wire.Herdr.SocketPath.value) ||
		wire.Herdr.TestedVersion.value != "herdr 0.9.1" {
		return Config{}, errors.New("host config herdr entry is invalid")
	}
	profiles, err := mapProfiles(*wire.Profiles)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Herdr: HerdrConfig{
			Path:          wire.Herdr.Path.value,
			SocketPath:    wire.Herdr.SocketPath.value,
			TestedVersion: wire.Herdr.TestedVersion.value,
		},
		Profiles: profiles,
	}, nil
}

func mapProfiles(wire []profileDTO) ([]profile.Profile, error) {
	if len(wire) != 0 && len(wire) != len(expectedProfiles) {
		return nil, fmt.Errorf("host config must declare exactly %d profiles", len(expectedProfiles))
	}
	profiles := make([]profile.Profile, len(wire))
	for index, candidate := range wire {
		expected := expectedProfiles[index]
		if !candidate.Key.present || !candidate.Label.present || !candidate.Provider.present || candidate.Environment == nil {
			return nil, errors.New("host config profile omits a required member")
		}
		if candidate.Key.value != expected.key {
			return nil, fmt.Errorf("host config profile %d must be %q", index, expected.key)
		}
		provider, err := profile.ParseProvider(candidate.Provider.value)
		if err != nil {
			return nil, fmt.Errorf("host config profile %s provider is invalid", candidate.Key.value)
		}
		if provider != expected.provider {
			return nil, fmt.Errorf("host config profile %s must use provider %s", candidate.Key.value, expected.provider)
		}
		environment, err := mapEnvironment(*candidate.Environment)
		if err != nil {
			return nil, err
		}
		profiles[index] = profile.Profile{
			Key:         profile.Key(candidate.Key.value),
			Label:       candidate.Label.value,
			Provider:    provider,
			Environment: environment,
		}
	}
	validated, err := profile.Validate(profiles)
	if err != nil {
		return nil, fmt.Errorf("validate host profiles: %w", err)
	}
	return validated, nil
}

func mapEnvironment(wire []environmentVariableDTO) ([]profile.EnvironmentVariable, error) {
	environment := make([]profile.EnvironmentVariable, len(wire))
	for index, candidate := range wire {
		if !candidate.Name.present || !candidate.Value.present {
			return nil, errors.New("host config environment entry omits a required member")
		}
		environment[index] = profile.EnvironmentVariable{Name: candidate.Name.value, Value: candidate.Value.value}
	}
	return environment, nil
}

func validAbsolutePath(value string) bool {
	return filepath.IsAbs(value) && filepath.Clean(value) == value && utf8.ValidString(value) && !strings.ContainsRune(value, 0)
}
