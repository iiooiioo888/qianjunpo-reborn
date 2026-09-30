package timedilation

import (
	"testing"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/degrade"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadsample"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/replay"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

func TestOverloadDrillDefaultScenario(t *testing.T) {
	clock := NewManualClock(time.Unix(0, 0))
	report := RunOverloadDrill(clock, DefaultOverloadSegments())
	if err := report.Pass(DefaultDrillEnvelope()); err != nil {
		t.Fatalf("drill failed: %v\n%s", err, FormatDrillLog(report, 15))
	}
}

func TestOverloadDrillTable(t *testing.T) {
	env := DefaultDrillEnvelope()
	cases := []struct {
		name     string
		segments []QueueSegment
		wantFail bool
	}{
		{name: "default", segments: DefaultOverloadSegments()},
		{
			name: "spike_only",
			segments: []QueueSegment{
				{LowWatermark, 3},
				{HighWatermark, 60},
				{LowWatermark, 150},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clock := NewManualClock(time.Unix(0, 0))
			report := RunOverloadDrill(clock, tc.segments)
			err := report.Pass(env)
			if tc.wantFail {
				if err == nil {
					t.Fatal("expected failure")
				}
				return
			}
			if err != nil {
				t.Fatalf("%v: %s", err, report.Summary())
			}
			if report.MaxObserved > CatchUpCapRate {
				t.Fatalf("max rate %v", report.MaxObserved)
			}
		})
	}
}

func TestDegradeRecoversAfterCooldown(t *testing.T) {
	clock := NewManualClock(time.Unix(0, 0))
	loop := NewControlLoop(clock)
	for i := 0; i < 40; i++ {
		loop.Tick(loadsample.Sample{QueueLength: HighWatermark})
		clock.Advance(100 * time.Millisecond)
	}
	if loop.Degrade.Current < degrade.L2 {
		t.Fatalf("expected escalation, got %v", loop.Degrade.Current)
	}
	for i := 0; i < 200; i++ {
		loop.Tick(loadsample.Sample{QueueLength: LowWatermark})
		clock.Advance(100 * time.Millisecond)
		if loop.Degrade.Current == degrade.L0 {
			return
		}
	}
	t.Fatalf("degrade stuck at %v", loop.Degrade.Current)
}

func TestOverloadDrillReplayHashUnchanged(t *testing.T) {
	clock := NewManualClock(time.Unix(0, 0))
	report := RunOverloadDrill(clock, DefaultOverloadSegments())

	sched := tactical.DemoSchedule()
	target := tactical.DemoTargetFrame()
	m := tactical.NewMatch(0xcafe)
	finalBase := m.RunSchedule(sched, target)

	recording := m.Recording()
	recording.TimeFlowRates = make([]uint32, len(report.Rates))
	for i, r := range report.Rates {
		recording.TimeFlowRates[i] = uint32(r)
	}

	got, err := tactical.ReplayFromRecording(recording)
	if err != nil {
		t.Fatal(err)
	}
	if got != finalBase {
		t.Fatalf("replay with rates: got %016x want %016x", got, finalBase)
	}

	gz, err := replay.MarshalGzip(recording)
	if err != nil {
		t.Fatal(err)
	}
	back, err := replay.UnmarshalGzip(gz)
	if err != nil {
		t.Fatal(err)
	}
	if len(back.TimeFlowRates) != len(report.Rates) {
		t.Fatalf("rates len=%d want %d", len(back.TimeFlowRates), len(report.Rates))
	}
	if replay.ComputeChainHash(back) != replay.ComputeChainHash(recording) {
		t.Fatal("chain hash must ignore time_flow_rate sidecar")
	}
}
