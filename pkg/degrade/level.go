// Package degrade implements L0–L5 degradation and ordered recovery for overload.
package degrade

import "fmt"

// Level is the degradation tier (L0 = healthy, L5 = most degraded).
type Level int

const (
	L0 Level = 0
	L1 Level = 1
	L2 Level = 2
	L3 Level = 3
	L4 Level = 4
	L5 Level = 5
)

func (l Level) String() string { return fmt.Sprintf("L%d", l) }

// Machine tracks current degradation and enforces monotonic recovery L5→L0.
type Machine struct {
	Current Level
}

// NewMachine starts at L0.
func NewMachine() *Machine { return &Machine{Current: L0} }

// Escalate moves one step toward L5 when triggers fire.
func (m *Machine) Escalate() Level {
	if m.Current < L5 {
		m.Current++
	}
	return m.Current
}

// EscalateTo jumps to at least the given level (for severe overload).
func (m *Machine) EscalateTo(min Level) Level {
	if min > m.Current {
		m.Current = min
	}
	if m.Current > L5 {
		m.Current = L5
	}
	return m.Current
}

// Recover steps down one level; ordered recovery only (never skip levels).
func (m *Machine) Recover() Level {
	if m.Current > L0 {
		m.Current--
	}
	return m.Current
}

// CanRecover reports whether another recover step is allowed.
func (m *Machine) CanRecover() bool { return m.Current > L0 }

// LevelFromLoad maps queue length and flow rate to suggested degradation level.
func LevelFromLoad(queueLen int, rateMilli int64) Level {
	switch {
	case queueLen >= 2000 || rateMilli <= 500:
		return L5
	case queueLen >= 1500 || rateMilli <= 700:
		return L4
	case queueLen >= 1200 || rateMilli <= 850:
		return L3
	case queueLen >= 1000 || rateMilli <= 950:
		return L2
	case queueLen >= 800 || rateMilli <= 9900:
		return L1
	default:
		return L0
	}
}

// SyncSuggested moves toward suggested level: escalate immediately, recover at most one level per call.
func (m *Machine) SyncSuggested(suggested Level) {
	for suggested > m.Current {
		m.Escalate()
	}
	if suggested < m.Current {
		m.Recover()
	}
}
