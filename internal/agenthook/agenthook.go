package agenthook

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/NielsdaWheelz/skidbladnir/internal/agentruntime"
	"github.com/NielsdaWheelz/skidbladnir/internal/herdr"
	"github.com/NielsdaWheelz/skidbladnir/internal/process"
)

const (
	maximumProcessAncestry = 128
	PublicationDeadline    = 3 * time.Second
	hookSessionStart       = "SessionStart"
)

var (
	ErrInvocationRejected = errors.New("unsupported provider hook event")
	ErrProviderInput      = errors.New("invalid provider hook input")
)

type Config struct {
	Herdr      *herdr.Client
	SocketPath string
	Profiles   []agentruntime.Profile
}

type invocation struct {
	provider agentruntime.Provider
}

// Run publishes only process-bound identity. It is a best-effort SessionStart
// observer and must never delay or block the provider's own startup.
func Run(parent context.Context, config Config, prepared Prepared) error {
	ctx, cancel := context.WithTimeout(parent, PublicationDeadline)
	defer cancel()
	if config.Herdr == nil || config.SocketPath == "" || os.Getenv("HERDR_SOCKET_PATH") != config.SocketPath {
		return nil
	}
	paneID := os.Getenv("HERDR_PANE_ID")
	if !validPaneID(paneID) {
		return nil
	}
	var pane struct {
		Type string `json:"type"`
		Pane struct {
			ID         string `json:"pane_id"`
			TerminalID string `json:"terminal_id"`
		} `json:"pane"`
	}
	if err := config.Herdr.Call(ctx, "pane.get", map[string]string{"pane_id": paneID}, &pane); err != nil ||
		pane.Type != "pane_info" || pane.Pane.ID != paneID || pane.Pane.TerminalID == "" {
		return errors.New("read herdr pane identity")
	}
	var processes struct {
		Type string `json:"type"`
		Info struct {
			PaneID         string  `json:"pane_id"`
			ForegroundPGID *uint32 `json:"foreground_process_group_id"`
			Foreground     []struct {
				PID uint32 `json:"pid"`
			} `json:"foreground_processes"`
		} `json:"process_info"`
	}
	if err := config.Herdr.Call(ctx, "pane.process_info", map[string]string{"pane_id": paneID}, &processes); err != nil ||
		processes.Type != "pane_process_info" || processes.Info.PaneID != paneID || processes.Info.ForegroundPGID == nil {
		return errors.New("read herdr pane process")
	}
	ancestry, err := process.ObserveAncestry(process.PID(os.Getpid()), maximumProcessAncestry)
	if err != nil {
		return errors.New("observe provider process")
	}
	pids := make([]process.PID, 0, len(processes.Info.Foreground))
	for _, foreground := range processes.Info.Foreground {
		pids = append(pids, process.PID(foreground.PID))
	}
	origin, found := agentruntime.HookOrigin(config.Profiles, ancestry, process.PID(*processes.Info.ForegroundPGID), pids)
	if !found || origin.Provider != prepared.invocation.provider {
		return nil
	}
	current, err := process.Observe(origin.PID)
	if err != nil || current.StartIdentity != origin.StartIdentity {
		return nil
	}
	profile, _ := agentruntime.MatchProfileEnvironment(config.Profiles, prepared.invocation.provider, os.LookupEnv)
	registration, err := agentruntime.EncodeRegistration(origin, profile, prepared.providerSessionID)
	if err != nil {
		return errors.New("encode agent runtime registration")
	}
	var report struct {
		Type string `json:"type"`
	}
	if err := config.Herdr.Call(ctx, "pane.report_metadata", map[string]any{"pane_id": paneID, "source": "user:skidbladnir",
		"tokens": map[string]string{agentruntime.RegistrationToken: registration}}, &report); err != nil || report.Type != "ok" {
		return errors.New("publish herdr pane identity")
	}
	return nil
}

func parseInvocation(providerText, eventText string) (invocation, error) {
	provider, err := agentruntime.ParseProvider(providerText)
	if err != nil || eventText != hookSessionStart {
		return invocation{}, ErrInvocationRejected
	}
	return invocation{provider: provider}, nil
}

func validPaneID(value string) bool {
	if len(value) == 0 || len(value) > 128 {
		return false
	}
	for _, symbol := range value {
		if symbol >= 'a' && symbol <= 'z' || symbol >= 'A' && symbol <= 'Z' ||
			symbol >= '0' && symbol <= '9' || symbol == ':' || symbol == '_' || symbol == '-' {
			continue
		}
		return false
	}
	return true
}
