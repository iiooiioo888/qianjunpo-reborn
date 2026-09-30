package roma

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/replay"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

func TestExportRecordingMatchesCLI(t *testing.T) {
	store := NewStore(nil)
	b, err := store.Join("default", 0)
	if err != nil {
		t.Fatal(err)
	}
	sched := tactical.DemoSchedule()
	m := b.Match
	final := m.RunSchedule(sched, tactical.DemoTargetFrame())
	rec, err := store.ExportRecording(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.FinalHash != final {
		t.Fatalf("final hash %016x != %016x", rec.FinalHash, final)
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
	if err != nil || got != final {
		t.Fatalf("verify replay: %016x err=%v", got, err)
	}
}
