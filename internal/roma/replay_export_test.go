package roma

import (
	"fmt"
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
	got, err := replayLiveJoinRecording(back)
	if err != nil || got != final {
		t.Fatalf("verify replay: %016x err=%v", got, err)
	}
}

// replayLiveJoinRecording mirrors tactical.ReplayFromRecording for Join's control-point duels.
func replayLiveJoinRecording(rec replay.Recording) (uint64, error) {
	if err := replay.VerifyConsistency(rec); err != nil {
		return 0, err
	}
	m := tactical.NewMatchWithControlPoints(rec.RNGSeed.S0, liveJoinControlPoints)
	m.RNG.SetState(rec.RNGSeed)
	if m.StateHash() != rec.InitialHash {
		return 0, fmt.Errorf("tactical: initial hash mismatch: got %016x want %016x", m.StateHash(), rec.InitialHash)
	}
	sched := make([]tactical.ScheduledCommand, 0, len(rec.Frames))
	for _, fc := range rec.Frames {
		cmd, err := tactical.Decode(fc.Payload)
		if err != nil {
			return 0, err
		}
		sched = append(sched, tactical.ScheduledCommand{SubmitFrame: fc.Frame, Cmd: cmd})
	}
	final := m.RunSchedule(sched, tactical.DemoTargetFrame())
	if final != rec.FinalHash {
		return final, fmt.Errorf("tactical: final hash mismatch: got %016x want %016x", final, rec.FinalHash)
	}
	return final, nil
}
