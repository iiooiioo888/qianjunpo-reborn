package integration

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/replay"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

func TestPhase2TacticalVerticalSlice(t *testing.T) {
	sched := tactical.DemoSchedule()
	m := tactical.NewMatch(0x706861736532)
	final := m.RunSchedule(sched, tactical.DemoTargetFrame())
	rec := m.Recording()
	if err := replay.VerifyTerminal(rec, final); err != nil {
		t.Fatal(err)
	}
	gz, err := replay.MarshalGzip(rec)
	if err != nil {
		t.Fatal(err)
	}
	back, err := replay.UnmarshalGzip(gz)
	if err != nil {
		t.Fatal(err)
	}
	got, err := tactical.ReplayFromRecording(back)
	if err != nil {
		t.Fatal(err)
	}
	if got != final {
		t.Fatalf("final hash mismatch")
	}
}
