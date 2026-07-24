package tools

import "env/engine"

// GaugeMux emits intermediate progress rows before replay completes.
type GaugeMux struct {
	Phase string
}

func (g *GaugeMux) Label(rows []engine.RowVec) string {
	if g.Phase == "early" {
		return "steady"
	}
	pairs := engine.PairMap(rows)
	if pairs["bind"] != pairs["bind_echo"] {
		return "pending"
	}
	if pairs["upper"] != pairs["upper_echo"] {
		return "pending"
	}
	if pairs["worker"] != pairs["worker_echo"] {
		return "pending"
	}
	return "steady"
}
