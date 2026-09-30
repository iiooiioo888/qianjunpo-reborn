package timedilation

import "time"

// Clock supplies controllable time for tests and production wall clocks.
type Clock interface {
	Now() time.Time
}

// WallClock uses time.Now.
type WallClock struct{}

func (WallClock) Now() time.Time { return time.Now() }

// ManualClock advances only when Stepped; used in unit tests.
type ManualClock struct {
	t time.Time
}

// NewManualClock starts at unix zero unless seed is non-zero.
func NewManualClock(start time.Time) *ManualClock {
	return &ManualClock{t: start}
}

func (c *ManualClock) Now() time.Time { return c.t }

// Advance moves the clock forward by d.
func (c *ManualClock) Advance(d time.Duration) { c.t = c.t.Add(d) }

// Layer identifies which logical clock tier is being adjusted.
type Layer int

const (
	LayerReal Layer = iota
	LayerRegion
	LayerActor
	LayerBattle
)

// LayerClocks holds cumulative time at each layer for one scope (e.g. a region).
type LayerClocks struct {
	RealNanos   int64
	RegionNanos int64
	ActorNanos  int64
	// BattleFrame is authoritative sim frame count at fixed 10fps; never dilated.
	BattleFrame uint64
}

// TickLayerClocks advances layered clocks for one control-loop step.
// battleAlways advances by exactly one lockstep frame worth of logical time (100ms nominal).
func TickLayerClocks(cl *LayerClocks, realDelta time.Duration, regionRate, actorRate Rate) {
	dt := realDelta.Nanoseconds()
	if dt < 0 {
		dt = 0
	}
	cl.RealNanos += dt
	regionDelta := dt * int64(regionRate) / RateScale
	cl.RegionNanos += regionDelta
	actorDelta := regionDelta * int64(actorRate) / RateScale
	cl.ActorNanos += actorDelta
	cl.BattleFrame++
}
