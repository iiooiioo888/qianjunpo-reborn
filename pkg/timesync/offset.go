package timesync

import "time"

// WallOffsetEstimator smooths Cristian offset and RTT from heartbeat samples.
type WallOffsetEstimator struct {
	alpha    float64
	offsetMs int64
	rtt      time.Duration
	samples  int
}

// NewWallOffsetEstimator creates an estimator. alpha is EWMA weight in (0,1]; default 0.25 if invalid.
func NewWallOffsetEstimator(alpha float64) *WallOffsetEstimator {
	if alpha <= 0 || alpha > 1 {
		alpha = 0.25
	}
	return &WallOffsetEstimator{alpha: alpha}
}

// Observe ingests one heartbeat sample.
func (e *WallOffsetEstimator) Observe(h HeartbeatSample) {
	off := h.CristianOffsetMs()
	rtt := h.RTT()
	if e.samples == 0 {
		e.offsetMs = off
		e.rtt = rtt
	} else {
		e.offsetMs = ewmaInt64(e.offsetMs, off, e.alpha)
		e.rtt = ewmaDuration(e.rtt, rtt, e.alpha)
	}
	e.samples++
}

// OffsetMs returns smoothed server−client wall skew in milliseconds.
func (e *WallOffsetEstimator) OffsetMs() int64 { return e.offsetMs }

// RTT returns smoothed round-trip time.
func (e *WallOffsetEstimator) RTT() time.Duration { return e.rtt }

// Samples returns how many observations were applied.
func (e *WallOffsetEstimator) Samples() int { return e.samples }

// ServerWallMs maps a client wall clock reading to estimated server wall time.
func (e *WallOffsetEstimator) ServerWallMs(clientWallMs int64) int64 {
	return clientWallMs + e.offsetMs
}

func ewmaInt64(prev, next int64, alpha float64) int64 {
	return int64(float64(prev)*(1-alpha) + float64(next)*alpha)
}

func ewmaDuration(prev, next time.Duration, alpha float64) time.Duration {
	return time.Duration(float64(prev)*(1-alpha) + float64(next)*alpha)
}
