package timesync

import (
	"errors"
	"time"
)

// DualTime pairs wall clock with deterministic simulation tick.
type DualTime struct {
	WallUnixMs int64
	SimTick    int64
}

// Equal reports whether wall and sim match.
func (d DualTime) Equal(o DualTime) bool {
	return d.WallUnixMs == o.WallUnixMs && d.SimTick == o.SimTick
}

// Clock advances sim tick monotonically; wall time comes from injectable source.
type Clock struct {
	now  func() time.Time
	tick int64
}

// NewClock builds a clock with optional time source (defaults to time.Now).
func NewClock(now func() time.Time) *Clock {
	if now == nil {
		now = time.Now
	}
	return &Clock{now: now}
}

// Now returns current dual timestamp.
func (c *Clock) Now() DualTime {
	return DualTime{WallUnixMs: c.now().UnixMilli(), SimTick: c.tick}
}

// AdvanceSim increments simulation tick (lockstep frame boundary).
func (c *Clock) AdvanceSim(steps int64) {
	if steps < 0 {
		return
	}
	c.tick += steps
}

// ZoneID identifies a Roma partition.
type ZoneID string

// Mapper translates sim time across zone boundaries (skeleton ratios).
type Mapper struct {
	// OffsetTicks maps destination zone sim offset from source at handoff.
	OffsetTicks map[ZoneID]int64
}

// MapCrossZone returns destination sim tick for a unit entering dst from src.
func (m *Mapper) MapCrossZone(src, dst ZoneID, srcSimTick int64) (int64, error) {
	if m == nil || m.OffsetTicks == nil {
		return 0, errors.New("timesync: mapper not configured")
	}
	off, ok := m.OffsetTicks[dst]
	if !ok {
		return 0, errors.New("timesync: unknown destination zone")
	}
	_ = src
	return srcSimTick + off, nil
}

// MapCrossZoneDual maps sim tick and preserves wall time at handoff (wall is not remapped).
func (m *Mapper) MapCrossZoneDual(src, dst ZoneID, at DualTime) (DualTime, error) {
	sim, err := m.MapCrossZone(src, dst, at.SimTick)
	if err != nil {
		return DualTime{}, err
	}
	return DualTime{WallUnixMs: at.WallUnixMs, SimTick: sim}, nil
}

// FreezePolicy pauses sim advancement while a cross-zone move is in flight.
type FreezePolicy struct {
	frozen bool
	hold   DualTime
}

// BeginCrossZoneMove freezes sim clock at the captured dual time.
func (f *FreezePolicy) BeginCrossZoneMove(at DualTime) {
	f.frozen = true
	f.hold = at
}

// EndCrossZoneMove resumes advancement.
func (f *FreezePolicy) EndCrossZoneMove() {
	f.frozen = false
}

// Frozen reports whether sim tick should not advance.
func (f *FreezePolicy) Frozen() bool {
	return f.frozen
}

// Held returns the frozen dual timestamp.
func (f *FreezePolicy) Held() DualTime {
	return f.hold
}

// AdvanceIfAllowed increments tick on clock when not frozen.
func (f *FreezePolicy) AdvanceIfAllowed(c *Clock, steps int64) {
	if f.Frozen() {
		return
	}
	c.AdvanceSim(steps)
}
