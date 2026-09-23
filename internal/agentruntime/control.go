package agentruntime

type Status struct {
	State  string `json:"state"`
	Source string `json:"source"`
	Reason string `json:"reason,omitempty"`
}

type Methods struct {
	Read      string `json:"read"`
	Send      string `json:"send"`
	Interrupt string `json:"interrupt"`
}

func (status Status) Valid() bool {
	switch status.State {
	case "working", "blocked", "idle", "unknown":
	default:
		return false
	}
	switch status.Source {
	case "herdr", "unavailable":
	default:
		return false
	}
	switch status.Reason {
	case "", "default_idle", "unrecognized", "observation_failed":
		return true
	default:
		return false
	}
}

func (methods Methods) Valid() bool {
	for _, method := range []string{methods.Read, methods.Send, methods.Interrupt} {
		switch method {
		case "terminal", "unavailable":
		default:
			return false
		}
	}
	return true
}
