package timedilation

import (
	"testing"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/degrade"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadsample"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/replay"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rng"
)

func TestControlLoopRecordsReplayRates(t *testing.T) {
	clock := NewManualClock(time.Unix(0, 0))
	loop := NewControlLoop(clock)
	rec := replay.NewRecorder(1, rng.State{S0: 1, S1: 1})
	for i := 0; i < 5; i++ {
		sample := loadsample.Sample{QueueLength: HighWatermark}
		if i > 2 {
			sample.QueueLength = LowWatermark
		}
		tick := loop.Tick(sample)
		rec.AddFrame(replay.FrameCommand{Frame: tick.BattleFrame, Player: 0, Payload: []byte{byte(i)}})
		rec.RecordTimeFlowRate(uint32(tick.Rate))
		clock.Advance(100 * time.Millisecond)
	}
	out := rec.Finish(99)
	if len(out.TimeFlowRates) != 5 {
		t.Fatalf("rates len=%d", len(out.TimeFlowRates))
	}
	if loop.Degrade.Current > degrade.L3 {
		t.Fatalf("degrade=%v", loop.Degrade.Current)
	}
}
