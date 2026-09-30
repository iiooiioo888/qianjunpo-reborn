// Package timedilation implements server-side time dilation control outside pure combat math.
//
// Boundary: simulation combat/board logic in pkg/sim, pkg/combat, pkg/fixed uses FP64 integers only.
// time_flow_rate scales how much wall-clock time maps to logical lockstep scheduling per region;
// it is recorded per tick for replay and never multiplies into FP64 combat formulas.
package timedilation

import "fmt"

// RateScale is parts-per-10000 where 10000 == 1.0x real-time flow.
const RateScale = 10000

// MinRate is the slowest normal dilation (0.1x).
const MinRate Rate = 1000

// MaxRate is full speed (1.0x).
const MaxRate Rate = RateScale

// CatchUpCapRate is the maximum temporary catch-up multiplier after recovery (1.5x).
const CatchUpCapRate Rate = 15000

// Rate is a fixed-point time flow multiplier (not used inside sim FP64 paths).
type Rate int64

// RateFromFloat builds a rate from a config float (initialization/tests only).
func RateFromFloat(v float64) Rate {
	return Rate(int64(v * RateScale))
}

// Float returns the rate as float64 (metrics/tests/logging only).
func (r Rate) Float() float64 {
	return float64(r) / RateScale
}

// String formats the rate for logs.
func (r Rate) String() string {
	return fmt.Sprintf("%.4f", r.Float())
}

// Clamp bounds r to [MinRate, CatchUpCapRate].
func (r Rate) Clamp() Rate {
	if r < MinRate {
		return MinRate
	}
	if r > CatchUpCapRate {
		return CatchUpCapRate
	}
	return r
}

// Lerp returns a linear blend between a and b by t in [0, RateScale].
func Lerp(a, b Rate, t int64) Rate {
	if t <= 0 {
		return a
	}
	if t >= RateScale {
		return b
	}
	delta := int64(b) - int64(a)
	return Rate(int64(a) + delta*t/RateScale)
}
