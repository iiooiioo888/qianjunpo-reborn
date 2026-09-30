// Package ai provides strategic (LLM cadence) and tactical (per-tick) AI skeletons.
package ai

import (
	"context"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/lockstep"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/sim"
)

// Order is a high-level strategic intent from the slow layer.
type Order struct {
	Verb     string
	TargetX  int8
	TargetY  int8
	IssuedAt time.Time
}

// StrategicPlanner produces orders on a ~5s cadence (mockable).
type StrategicPlanner interface {
	Plan(ctx context.Context, snap sim.Snapshot) (Order, error)
}

// MockPlanner returns deterministic orders for tests.
type MockPlanner struct {
	Cadence time.Duration
	Last    time.Time
	Seq     int
}

func (m *MockPlanner) Plan(_ context.Context, _ sim.Snapshot) (Order, error) {
	now := time.Now()
	if !m.Last.IsZero() && now.Sub(m.Last) < m.Cadence {
		return Order{}, nil
	}
	m.Last = now
	m.Seq++
	return Order{Verb: "advance", TargetX: int8(m.Seq % 3), TargetY: 1, IssuedAt: now}, nil
}

// TacticalExecutor turns orders into lockstep commands each tick.
type TacticalExecutor struct {
	PlayerID uint8
 pending  *Order
}

// NewTacticalExecutor binds a player id.
func NewTacticalExecutor(playerID uint8) *TacticalExecutor {
	return &TacticalExecutor{PlayerID: playerID}
}

// AcceptOrder stores the latest strategic order.
func (t *TacticalExecutor) AcceptOrder(o Order) {
	if o.Verb == "" {
		return
	}
	copy := o
	t.pending = &copy
}

// Tick emits at most one command per lockstep frame.
func (t *TacticalExecutor) Tick() (lockstep.CommandPacket, bool) {
	if t.pending == nil {
		return lockstep.CommandPacket{}, false
	}
	cmd := lockstep.CommandPacket{
		PlayerID: t.PlayerID,
		MoveX:    t.pending.TargetX,
		MoveY:    t.pending.TargetY,
	}
	t.pending = nil
	return cmd, true
}

// NPCFallback returns template dialogue/actions when edge inference is unavailable (<1s path).
func NPCFallback(persona string) string {
	switch persona {
	case "guard":
		return "Hold the line!"
	default:
		return "For the realm!"
	}
}

// Pipeline connects snapshot → strategic plan → tactical command.
type Pipeline struct {
	Planner  StrategicPlanner
	Tactical *TacticalExecutor
}

// RunStep executes one strategic+tactical cycle; returns command if tactical produced one.
func (p *Pipeline) RunStep(ctx context.Context, snap sim.Snapshot) (lockstep.CommandPacket, bool) {
	if p.Planner != nil {
		if o, err := p.Planner.Plan(ctx, snap); err == nil && o.Verb != "" {
			p.Tactical.AcceptOrder(o)
		}
	}
	return p.Tactical.Tick()
}
