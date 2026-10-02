package roma

import (
	"encoding/json"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

func TestTacticalViewSnapshotMatchesDemoSchedule(t *testing.T) {
	store := NewStore(nil)
	b, err := store.Join("default", 0)
	if err != nil {
		t.Fatal(err)
	}
	ref := tactical.NewMatch(b.Match.Seed)
	sched := tactical.DemoSchedule()
	idx := 0
	for frame := uint64(0); frame < tactical.DemoTargetFrame(); frame++ {
		for idx < len(sched) && sched[idx].SubmitFrame == frame {
			item := sched[idx]
			_, _, err := store.SubmitTacticalCommand(b.ID, uint32(item.Cmd.PlayerID), uint32(item.Cmd.Kind), item.Cmd.UnitID, int32(item.Cmd.To.X), int32(item.Cmd.To.Y), uint32(item.Cmd.SkillID))
			if err != nil {
				t.Fatalf("submit frame %d: %v", frame, err)
			}
			if err := ref.Submit(item.Cmd); err != nil {
				t.Fatalf("ref submit: %v", err)
			}
			idx++
		}
		_, _, _, _, err := store.StepLockstep(b.ID, 1)
		if err != nil {
			t.Fatal(err)
		}
		ref.StepLockstep()
	}

	raw, hash, frame, err := store.TacticalViewSnapshotJSON(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if hash != ref.StateHash() {
		t.Fatalf("hash svc=%016x ref=%016x", hash, ref.StateHash())
	}
	if frame != ref.Frame {
		t.Fatalf("frame svc=%d ref=%d", frame, ref.Frame)
	}

	var view tactical.ViewSnapshot
	if err := json.Unmarshal(raw, &view); err != nil {
		t.Fatal(err)
	}
	want := tactical.MatchToViewSnapshot(ref, view.TimeFlowRateParts)
	if view.LockstepFrame != want.LockstepFrame || view.InitialStateHash != want.InitialStateHash {
		t.Fatalf("meta frame=%d hash=%s want frame=%d hash=%s", view.LockstepFrame, view.InitialStateHash, want.LockstepFrame, want.InitialStateHash)
	}
	if len(view.Units) != len(want.Units) {
		t.Fatalf("units=%d want=%d", len(view.Units), len(want.Units))
	}
	assertUnitsAlign(t, view.Units, want.Units)
}

func assertUnitsAlign(t *testing.T, got, want []tactical.ViewUnit) {
	if len(got) != len(want) {
		t.Fatalf("units=%d want=%d", len(got), len(want))
	}
	byID := make(map[uint32]tactical.ViewUnit, len(want))
	for _, u := range want {
		byID[u.ID] = u
	}
	for _, u := range got {
		w, ok := byID[u.ID]
		if !ok {
			t.Fatalf("unexpected unit id %d", u.ID)
		}
		if u.X != w.X || u.Y != w.Y || u.HP != w.HP {
			t.Fatalf("unit %d: got %+v want %+v", u.ID, u, w)
		}
	}
}
