package timesync

import "errors"

// SimCadence optional wall-driven extrapolation between heartbeats (0 = hold last server sim).
type SimCadence struct {
	TicksPerSecond int64
}

// ExtrapolateSim advances anchor sim tick by elapsed server wall time at cadence.
func (c SimCadence) ExtrapolateSim(anchor DualTime, serverWallMs int64) int64 {
	if c.TicksPerSecond <= 0 {
		return anchor.SimTick
	}
	delta := serverWallMs - anchor.WallUnixMs
	if delta <= 0 {
		return anchor.SimTick
	}
	ticks := delta * c.TicksPerSecond / 1000
	return anchor.SimTick + ticks
}

// TickAligner tracks authoritative server sim from heartbeats and aligns local lockstep tick.
type TickAligner struct {
	cadence SimCadence
	anchor  DualTime
	offsetMs int64
	hasAnchor bool
}

// NewTickAligner builds an aligner with optional cadence for between-heartbeat extrapolation.
func NewTickAligner(cadence SimCadence) *TickAligner {
	return &TickAligner{cadence: cadence}
}

// ObserveHeartbeat updates anchor from a heartbeat and wall offset estimator.
func (a *TickAligner) ObserveHeartbeat(h HeartbeatSample, offset *WallOffsetEstimator) {
	a.anchor = h.Server
	if offset != nil {
		a.offsetMs = offset.OffsetMs()
	} else {
		a.offsetMs = h.CristianOffsetMs()
	}
	a.hasAnchor = true
}

// EstimatedServerSim returns server sim tick at clientWallMs using anchor + offset + cadence.
func (a *TickAligner) EstimatedServerSim(clientWallMs int64) (int64, error) {
	if !a.hasAnchor {
		return 0, errors.New("timesync: tick aligner has no anchor")
	}
	serverWall := clientWallMs + a.offsetMs
	return a.cadence.ExtrapolateSim(a.anchor, serverWall), nil
}

// AlignClientTick clamps local sim to stay within [server−maxLag, server+maxLead].
func AlignClientTick(localSim, estimatedServerSim, maxLead, maxLag int64) (aligned int64, clamped bool) {
	if maxLead < 0 {
		maxLead = 0
	}
	if maxLag < 0 {
		maxLag = 0
	}
	hi := estimatedServerSim + maxLead
	lo := estimatedServerSim - maxLag
	aligned = localSim
	if aligned > hi {
		aligned = hi
		clamped = true
	}
	if aligned < lo {
		aligned = lo
		clamped = true
	}
	return aligned, clamped
}

// TickBoundary returns the next sim tick boundary for scheduling inputs (lockstep frame).
func TickBoundary(currentSim, step int64) int64 {
	if step <= 0 {
		step = 1
	}
	if currentSim%step == 0 {
		return currentSim
	}
	return currentSim + (step - currentSim%step)
}
