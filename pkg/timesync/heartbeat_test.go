package timesync

import (
	"testing"
	"time"
)

func TestHeartbeatRTTAndCristianOffset(t *testing.T) {
	clientSent := DualTime{WallUnixMs: 1000, SimTick: 5}
	server := DualTime{WallUnixMs: 1050, SimTick: 7}
	h := NewHeartbeatSample(clientSent, server, 1100)

	if got := h.RTT(); got != 100*time.Millisecond {
		t.Fatalf("rtt=%v want 100ms", got)
	}
	// mid client = 1000 + 50 = 1050, offset = 1050 - 1050 = 0
	if off := h.CristianOffsetMs(); off != 0 {
		t.Fatalf("offset=%d want 0", off)
	}

	serverAhead := DualTime{WallUnixMs: 1080, SimTick: 7}
	h2 := NewHeartbeatSample(clientSent, serverAhead, 1100)
	if off := h2.CristianOffsetMs(); off != 30 {
		t.Fatalf("offset=%d want 30", off)
	}
}

func TestWallOffsetEstimatorSmoothing(t *testing.T) {
	est := NewWallOffsetEstimator(1) // no smoothing for deterministic test
	h := NewHeartbeatSample(
		DualTime{WallUnixMs: 0, SimTick: 0},
		DualTime{WallUnixMs: 150, SimTick: 0},
		200,
	)
	est.Observe(h)
	if est.OffsetMs() != 50 {
		t.Fatalf("offset=%d want 50", est.OffsetMs())
	}
	if est.RTT() != 200*time.Millisecond {
		t.Fatalf("rtt=%v", est.RTT())
	}
	if est.ServerWallMs(1000) != 1050 {
		t.Fatalf("server wall=%d", est.ServerWallMs(1000))
	}
}

func TestTickAlignerAndCadence(t *testing.T) {
	aligner := NewTickAligner(SimCadence{TicksPerSecond: 10})
	est := NewWallOffsetEstimator(1)
	h := NewHeartbeatSample(
		DualTime{WallUnixMs: 1000, SimTick: 100},
		DualTime{WallUnixMs: 1050, SimTick: 110},
		1100,
	)
	aligner.ObserveHeartbeat(h, est)

	sim, err := aligner.EstimatedServerSim(1500)
	if err != nil {
		t.Fatal(err)
	}
	// server wall at client 1500: 1500 + 0 offset; delta 450ms from anchor wall 1050 -> 4 ticks at 10/s
	if sim != 110+4 {
		t.Fatalf("sim=%d want %d", sim, 114)
	}

	aligned, clamped := AlignClientTick(200, 110, 2, 3)
	if !clamped || aligned != 112 {
		t.Fatalf("aligned=%d clamped=%v", aligned, clamped)
	}
}

func TestTickBoundary(t *testing.T) {
	if got := TickBoundary(10, 3); got != 12 {
		t.Fatalf("got %d want 12", got)
	}
	if got := TickBoundary(9, 3); got != 9 {
		t.Fatalf("got %d want 9", got)
	}
}

func TestMapCrossZoneDual(t *testing.T) {
	m := &Mapper{OffsetTicks: map[ZoneID]int64{"roma-b": 50}}
	dual, err := m.MapCrossZoneDual("roma-a", "roma-b", DualTime{WallUnixMs: 99, SimTick: 10})
	if err != nil || dual.WallUnixMs != 99 || dual.SimTick != 60 {
		t.Fatalf("dual=%+v err=%v", dual, err)
	}
}
