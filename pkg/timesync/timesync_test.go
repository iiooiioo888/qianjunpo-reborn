package timesync

import (
	"testing"
	"time"
)

func TestDualTimeMonotonicSim(t *testing.T) {
	fixed := time.Unix(1_700_000_000, 0)
	clk := NewClock(func() time.Time { return fixed })
	before := clk.Now()
	clk.AdvanceSim(10)
	after := clk.Now()
	if after.SimTick != before.SimTick+10 {
		t.Fatalf("sim tick=%d want %d", after.SimTick, before.SimTick+10)
	}
	if after.WallUnixMs != before.WallUnixMs {
		t.Fatal("wall should not change without time source change")
	}
}

func TestCrossZoneMapping(t *testing.T) {
	m := &Mapper{OffsetTicks: map[ZoneID]int64{"roma-b": 100}}
	dst, err := m.MapCrossZone("roma-a", "roma-b", 50)
	if err != nil || dst != 150 {
		t.Fatalf("dst=%d err=%v", dst, err)
	}
}

func TestFreezeOnCrossZoneMove(t *testing.T) {
	clk := NewClock(time.Now)
	fp := &FreezePolicy{}
	start := clk.Now()
	fp.BeginCrossZoneMove(start)
	fp.AdvanceIfAllowed(clk, 5)
	if clk.Now().SimTick != start.SimTick {
		t.Fatal("sim should not advance while frozen")
	}
	fp.EndCrossZoneMove()
	fp.AdvanceIfAllowed(clk, 5)
	if clk.Now().SimTick != start.SimTick+5 {
		t.Fatal("sim should advance after unfreeze")
	}
}
