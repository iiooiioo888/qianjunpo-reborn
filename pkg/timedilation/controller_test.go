package timedilation

import (
	"testing"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadsample"
)

func TestTargetRateWatermarks(t *testing.T) {
	if TargetRate(0) != MaxRate {
		t.Fatalf("low watermark: got %v", TargetRate(0))
	}
	if TargetRate(LowWatermark) != MaxRate {
		t.Fatalf("at low: got %v", TargetRate(LowWatermark))
	}
	if TargetRate(HighWatermark) != MinRate {
		t.Fatalf("high: got %v", TargetRate(HighWatermark))
	}
	mid := TargetRate(600)
	if mid >= MaxRate || mid <= MinRate {
		t.Fatalf("mid lerp out of range: %v", mid)
	}
}

func TestSlowdownDuration(t *testing.T) {
	clock := NewManualClock(time.Unix(0, 0))
	cfg := DefaultControllerConfig()
	c := NewController(cfg, clock, nil)
	hot := loadsample.Sample{QueueLength: HighWatermark}

	const step = 100 * time.Millisecond
	var elapsed time.Duration
	for elapsed < 8*time.Second {
		c.Step(hot, uint64(elapsed/step))
		clock.Advance(step)
		elapsed += step
		if c.Rate <= MinRate+50 {
			break
		}
	}
	got := elapsed.Seconds()
	if got < 3.8 || got > 5.2 {
		t.Fatalf("1.0→0.1 expected ~4.5s, got %.2fs rate=%s", got, c.Rate)
	}
}

func TestRecoveryDuration(t *testing.T) {
	clock := NewManualClock(time.Unix(0, 0))
	cfg := DefaultControllerConfig()
	c := NewController(cfg, clock, nil)
	c.Rate = MinRate
	cool := loadsample.Sample{QueueLength: LowWatermark}

	const step = 100 * time.Millisecond
	var elapsed time.Duration
	for elapsed < 20*time.Second {
		c.Step(cool, uint64(elapsed/step))
		clock.Advance(step)
		elapsed += step
		if c.Rate >= MaxRate-50 {
			break
		}
	}
	got := elapsed.Seconds()
	if got < 10.0 || got > 13.5 {
		t.Fatalf("0.1→1.0 expected ~11.6s, got %.2fs rate=%s", got, c.Rate)
	}
}

func TestCatchUpCap(t *testing.T) {
	clock := NewManualClock(time.Unix(0, 0))
	c := NewController(DefaultControllerConfig(), clock, nil)
	c.BeginCatchUp()
	c.Rate = MaxRate
	for i := 0; i < 50; i++ {
		c.Step(loadsample.Sample{QueueLength: LowWatermark}, uint64(i))
		clock.Advance(100 * time.Millisecond)
	}
	if c.Rate > CatchUpCapRate {
		t.Fatalf("catch-up exceeded cap: %v", c.Rate)
	}
}

func TestLayerClocksBattleFixed(t *testing.T) {
	var cl LayerClocks
	d := 100 * time.Millisecond
	for i := 0; i < 10; i++ {
		TickLayerClocks(&cl, d, MinRate, MaxRate)
	}
	if cl.BattleFrame != 10 {
		t.Fatalf("battle frames=%d want 10", cl.BattleFrame)
	}
	if cl.RegionNanos >= cl.RealNanos {
		t.Fatalf("region should lag under slow rate")
	}
}
