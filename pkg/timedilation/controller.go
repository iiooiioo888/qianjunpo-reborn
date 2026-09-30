package timedilation

import (
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadsample"
)

// Watermarks for queue-length-driven target rate (high → slow down, low → speed up).
const (
	HighWatermark = 1000
	LowWatermark  = 200
)

// ControllerConfig tunes progressive slowdown/recovery.
type ControllerConfig struct {
	// SlowdownPerSec is how fast rate drops toward target when overloaded (parts/sec at RateScale).
	SlowdownPerSec int64
	// RecoveryPerSec is how fast rate rises when underloaded (slower than slowdown).
	RecoveryPerSec int64
	TickInterval   time.Duration
}

// DefaultControllerConfig targets ~4.5s from 1.0→0.1 and ~11.6s from 0.1→1.0 at 10Hz control ticks.
func DefaultControllerConfig() ControllerConfig {
	return ControllerConfig{
		SlowdownPerSec: 2000, // 0.2x per second toward min when hot
		RecoveryPerSec: 776,  // ~0.0776x per second toward max when cool
		TickInterval:   100 * time.Millisecond,
	}
}

// Controller adjusts region time_flow_rate from load samples.
type Controller struct {
	cfg       ControllerConfig
	clock     Clock
	predictor loadsample.Predictor
	Rate      Rate
	CatchUp   bool
	lastTick  time.Time
	TickLog   []TickRecord
}

// TickRecord stores per-tick rate for replay compatibility.
type TickRecord struct {
	BattleFrame uint64
	Rate        Rate
	Layer       Layer
}

// NewController creates a controller at full rate 1.0.
func NewController(cfg ControllerConfig, clock Clock, pred loadsample.Predictor) *Controller {
	if pred == nil {
		pred = loadsample.NoOpPredictor{}
	}
	now := clock.Now()
	return &Controller{
		cfg:       cfg,
		clock:     clock,
		predictor: pred,
		Rate:      MaxRate,
		lastTick:  now,
	}
}

// TargetRate maps queue length to desired flow using watermarks and lerp.
func TargetRate(queueLen int) Rate {
	if queueLen >= HighWatermark {
		return MinRate
	}
	if queueLen <= LowWatermark {
		return MaxRate
	}
	span := int64(HighWatermark - LowWatermark)
	t := int64(queueLen-LowWatermark) * RateScale / span
	return Lerp(MaxRate, MinRate, t)
}

// Step applies one control tick from load sample and optional battle frame index.
func (c *Controller) Step(sample loadsample.Sample, battleFrame uint64) TickRecord {
	now := c.clock.Now()
	if c.lastTick.IsZero() {
		c.lastTick = now
	}
	dt := now.Sub(c.lastTick)
	if dt <= 0 {
		dt = c.cfg.TickInterval
	}
	c.lastTick = now

	predQ, _ := c.predictor.Predict(sample)
	queueForTarget := sample.QueueLength
	if predQ > queueForTarget {
		queueForTarget = predQ
	}
	target := TargetRate(queueForTarget)

	sec := dt.Seconds()
	if sec <= 0 {
		sec = c.cfg.TickInterval.Seconds()
	}
	slowDelta := Rate(int64(float64(c.cfg.SlowdownPerSec) * sec))
	recDelta := Rate(int64(float64(c.cfg.RecoveryPerSec) * sec))

	if c.Rate > target {
		c.Rate -= slowDelta
		if c.Rate < target {
			c.Rate = target
		}
	} else if c.Rate < target {
		c.Rate += recDelta
		if c.Rate > target {
			c.Rate = target
		}
	}

	if c.CatchUp && c.Rate < CatchUpCapRate {
		c.Rate += recDelta / 2
		if c.Rate > CatchUpCapRate {
			c.Rate = CatchUpCapRate
		}
	}
	c.Rate = c.Rate.Clamp()

	rec := TickRecord{BattleFrame: battleFrame, Rate: c.Rate, Layer: LayerRegion}
	c.TickLog = append(c.TickLog, rec)
	return rec
}

// BeginCatchUp enables post-recovery catch-up mode until rate returns to MaxRate.
func (c *Controller) BeginCatchUp() { c.CatchUp = true }

// EndCatchUp clears catch-up once stabilized at full rate.
func (c *Controller) EndCatchUp() {
	if c.Rate >= MaxRate {
		c.CatchUp = false
		c.Rate = MaxRate
	}
}
