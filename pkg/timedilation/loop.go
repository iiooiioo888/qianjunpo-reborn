package timedilation

import (
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/degrade"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadsample"
)

// ControlLoop wires dilation controller, degradation, layered clocks, and catch-up.
type ControlLoop struct {
	Controller *Controller
	Degrade    *degrade.Machine
	Clocks     LayerClocks
}

// NewControlLoop builds a loop with defaults.
func NewControlLoop(clock Clock) *ControlLoop {
	return &ControlLoop{
		Controller: NewController(DefaultControllerConfig(), clock, nil),
		Degrade:    degrade.NewMachine(),
	}
}

// Tick runs one 10fps-aligned control iteration.
func (l *ControlLoop) Tick(sample loadsample.Sample) TickRecord {
 lvl := degrade.LevelFromLoad(sample.QueueLength, int64(l.Controller.Rate))
	l.Degrade.SyncSuggested(lvl)
	rec := l.Controller.Step(sample, l.Clocks.BattleFrame)
	TickLayerClocks(&l.Clocks, DefaultControllerConfig().TickInterval, rec.Rate, MaxRate)
	if l.Degrade.Current == degrade.L0 && l.Controller.Rate < MaxRate {
		l.Controller.BeginCatchUp()
	}
	if l.Controller.CatchUp && l.Controller.Rate >= MaxRate {
		l.Controller.EndCatchUp()
	}
	return rec
}

// RunCatchUpTestHelper exposes catch-up cap behavior for integration tests.
func (l *ControlLoop) ApplyCatchUpFor(d time.Duration, clock *ManualClock) {
	deadline := clock.Now().Add(d)
	for clock.Now().Before(deadline) {
		l.Tick(loadsample.Sample{QueueLength: LowWatermark})
		clock.Advance(DefaultControllerConfig().TickInterval)
	}
}
