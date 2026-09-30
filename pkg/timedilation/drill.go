package timedilation

import (
	"fmt"
	"strings"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/degrade"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadsample"
)

// QueueSegment holds a fixed queue length for a number of control ticks (100ms each by default).
type QueueSegment struct {
	QueueLength int
	Ticks       int
}

// DrillEnvelope is the pass/fail window for overload drill assertions.
type DrillEnvelope struct {
	MinSlowdownSec float64
	MaxSlowdownSec float64
	MinRecoverySec float64
	MaxRecoverySec float64
	MinHeldMinRate time.Duration // time near MinRate during hot hold
}

// DefaultDrillEnvelope matches whitepaper targets (~4.5s slowdown, ~11.6s recovery).
func DefaultDrillEnvelope() DrillEnvelope {
	return DrillEnvelope{
		MinSlowdownSec: 3.8,
		MaxSlowdownSec: 5.2,
		MinRecoverySec: 7.5, // catch-up after L0 recovery can beat isolated 11.6s curve
		MaxRecoverySec: 13.5,
		MinHeldMinRate: 500 * time.Millisecond,
	}
}

// DrillReport captures observable outcomes from an overload scenario.
type DrillReport struct {
	Segments          []QueueSegment
	TickRecords       []TickRecord
	DegradePath       []degrade.Level
	Rates             []Rate
	SlowdownSec       float64
	RecoverySec       float64
	MinObserved       Rate
	MaxObserved       Rate
	MaxDegrade        degrade.Level
	ReachedCatchUp    bool
	CatchUpExceeded   bool
	Crashed           bool
	CrashReason       string
}

// Pass checks timing and rate bounds from the drill report.
func (r DrillReport) Pass(env DrillEnvelope) error {
	if r.Crashed {
		return fmt.Errorf("drill crashed: %s", r.CrashReason)
	}
	if r.MinObserved > MinRate+100 {
		return fmt.Errorf("never slowed below %s (min=%s)", MinRate, r.MinObserved)
	}
	if r.MaxObserved > CatchUpCapRate {
		return fmt.Errorf("rate exceeded catch-up cap: %s", r.MaxObserved)
	}
	if r.CatchUpExceeded {
		return fmt.Errorf("catch-up cap violated during drill")
	}
	if r.SlowdownSec < env.MinSlowdownSec || r.SlowdownSec > env.MaxSlowdownSec {
		return fmt.Errorf("slowdown %.2fs outside [%.1f, %.1f]", r.SlowdownSec, env.MinSlowdownSec, env.MaxSlowdownSec)
	}
	if r.RecoverySec < env.MinRecoverySec || r.RecoverySec > env.MaxRecoverySec {
		return fmt.Errorf("recovery %.2fs outside [%.1f, %.1f]", r.RecoverySec, env.MinRecoverySec, env.MaxRecoverySec)
	}
	if r.MaxDegrade < degrade.L1 {
		return fmt.Errorf("degrade never escalated past L0 (max=%v)", r.MaxDegrade)
	}
	return nil
}

// Summary returns a one-line human-readable result.
func (r DrillReport) Summary() string {
	return fmt.Sprintf(
		"slowdown=%.2fs recovery=%.2fs rate=[%s,%s] degrade_max=%s catch_up=%v",
		r.SlowdownSec, r.RecoverySec, r.MinObserved, r.MaxObserved, r.MaxDegrade, r.ReachedCatchUp,
	)
}

// DefaultOverloadSegments ramps load, holds hot, then cools — drives slowdown → hold → recovery → catch-up.
func DefaultOverloadSegments() []QueueSegment {
	return []QueueSegment{
		{QueueLength: LowWatermark, Ticks: 5},
		{QueueLength: 600, Ticks: 10},
		{QueueLength: HighWatermark, Ticks: 55},
		{QueueLength: LowWatermark, Ticks: 160},
	}
}

// RunOverloadDrill drives the control loop through segments on a manual clock (deterministic).
func RunOverloadDrill(clock *ManualClock, segments []QueueSegment) DrillReport {
	cfg := DefaultControllerConfig()
	loop := NewControlLoop(clock)
	report := DrillReport{Segments: segments, MinObserved: MaxRate, MaxDegrade: degrade.L0}

	var slowdownStart time.Time
	var recoveryStart time.Time
	var atMinSince time.Time
	var wasHot bool
	slowdownDone := false
	recoveryDone := false

	step := cfg.TickInterval
	for _, seg := range segments {
		if seg.QueueLength >= HighWatermark {
			wasHot = true
		}
		for i := 0; i < seg.Ticks; i++ {
			sample := loadsample.Sample{QueueLength: seg.QueueLength}
			rec := loop.Tick(sample)
			report.TickRecords = append(report.TickRecords, rec)
			report.Rates = append(report.Rates, rec.Rate)
			report.DegradePath = append(report.DegradePath, loop.Degrade.Current)
			if loop.Degrade.Current > report.MaxDegrade {
				report.MaxDegrade = loop.Degrade.Current
			}
			if rec.Rate < report.MinObserved {
				report.MinObserved = rec.Rate
			}
			if rec.Rate > report.MaxObserved {
				report.MaxObserved = rec.Rate
			}
			if loop.Controller.CatchUp {
				report.ReachedCatchUp = true
			}
			if rec.Rate > CatchUpCapRate {
				report.CatchUpExceeded = true
			}

			now := clock.Now()
			if !slowdownDone {
				if slowdownStart.IsZero() && seg.QueueLength >= HighWatermark-50 {
					slowdownStart = now
				}
				if rec.Rate <= MinRate+80 {
					if atMinSince.IsZero() {
						atMinSince = now
					}
				}
				if !atMinSince.IsZero() && now.Sub(atMinSince) >= DefaultDrillEnvelope().MinHeldMinRate {
					report.SlowdownSec = now.Sub(slowdownStart).Seconds()
					slowdownDone = true
				}
			}
			if !recoveryDone && wasHot && seg.QueueLength <= LowWatermark && rec.Rate < MaxRate-80 {
				if recoveryStart.IsZero() {
					recoveryStart = now
				}
			}
			if !recoveryDone && !recoveryStart.IsZero() && rec.Rate >= MaxRate-80 {
				report.RecoverySec = now.Sub(recoveryStart).Seconds()
				recoveryDone = true
			}

			clock.Advance(step)
		}
	}

	if report.SlowdownSec == 0 && slowdownStart.IsZero() == false {
		report.Crashed = true
		report.CrashReason = "never reached held min rate"
	}
	if report.RecoverySec == 0 && recoveryDone == false {
		report.Crashed = true
		if report.CrashReason == "" {
			report.CrashReason = "never recovered to full rate"
		}
	}
	return report
}

// FormatDrillLog prints tick samples for CLI debugging.
func FormatDrillLog(r DrillReport, every int) string {
	if every < 1 {
		every = 1
	}
	var b strings.Builder
	for i, rec := range r.TickRecords {
		if i%every != 0 && i != len(r.TickRecords)-1 {
			continue
		}
		lvl := degrade.L0
		if i < len(r.DegradePath) {
			lvl = r.DegradePath[i]
		}
		fmt.Fprintf(&b, "tick %4d frame=%4d rate=%s degrade=%s\n", i, rec.BattleFrame, rec.Rate, lvl)
	}
	return b.String()
}
