package roma

import (
	"encoding/json"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

// Live export path: TacticalViewSnapshotJSON must surface match outcome after lockstep victory.
func TestTacticalViewSnapshotJSONReportsWipeout(t *testing.T) {
	store := NewStore(nil)
	b, err := store.Join("default", 0)
	if err != nil {
		t.Fatal(err)
	}
	enemy := b.Match.Units[tactical.UnitIDPlayer1]
	enemy.Stats.HP = enemy.Stats.HP.Sub(enemy.Stats.HP)
	b.Match.Board.ClearUnit(enemy.Pos)
	_, _, finished, winner, _, err := store.StepLockstep(b.ID, 1, StepLockstepOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if !finished || winner != 0 {
		t.Fatalf("step finished=%v winner=%d", finished, winner)
	}

	raw, _, _, err := store.TacticalViewSnapshotJSON(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	var view tactical.ViewSnapshot
	if err := json.Unmarshal(raw, &view); err != nil {
		t.Fatal(err)
	}
	if view.EndReason != "wipeout" {
		t.Fatalf("endReason=%q", view.EndReason)
	}
	if view.Winner == nil || *view.Winner != 0 {
		t.Fatalf("winner=%v", view.Winner)
	}
}

func TestTacticalViewSnapshotJSONReportsOccupy(t *testing.T) {
	store := NewStore(nil)
	b, err := store.Join("default", 0)
	if err != nil {
		t.Fatal(err)
	}
	b.Match = tactical.NewMatchWithControlPoints(b.Match.Seed, []tactical.ControlPoint{
		{Pos: board.Coord{2, 8}, HoldFrames: 2},
	})
	for i := 0; i < 2 && !b.Match.Finished; i++ {
		if _, _, _, _, _, err := store.StepLockstep(b.ID, 1, StepLockstepOpts{}); err != nil {
			t.Fatal(err)
		}
	}
	if !b.Match.Finished || b.Match.EndReason != tactical.EndCapture {
		t.Fatalf("match finished=%v reason=%d", b.Match.Finished, b.Match.EndReason)
	}

	raw, _, _, err := store.TacticalViewSnapshotJSON(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	var view tactical.ViewSnapshot
	if err := json.Unmarshal(raw, &view); err != nil {
		t.Fatal(err)
	}
	if view.EndReason != "occupy" {
		t.Fatalf("endReason=%q", view.EndReason)
	}
	if view.Winner == nil || *view.Winner != 0 {
		t.Fatalf("winner=%v", view.Winner)
	}
}

func TestTacticalViewSnapshotJSONReportsTimeout(t *testing.T) {
	store := NewStore(nil)
	b, err := store.Join("default", 0)
	if err != nil {
		t.Fatal(err)
	}
	b.Match.Units[tactical.UnitIDPlayer0].Stats.HP = fixed.FromInt(50)
	b.Match.Units[tactical.UnitIDPlayer1].Stats.HP = fixed.FromInt(30)
	driveStoreTimeoutPass(t, store, b.ID)
	if b.Match.EndReason != tactical.EndTimeout || b.Match.Winner != 0 {
		t.Fatalf("match reason=%d winner=%d", b.Match.EndReason, b.Match.Winner)
	}

	raw, _, _, err := store.TacticalViewSnapshotJSON(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	var view tactical.ViewSnapshot
	if err := json.Unmarshal(raw, &view); err != nil {
		t.Fatal(err)
	}
	if view.EndReason != "timeout" {
		t.Fatalf("endReason=%q", view.EndReason)
	}
	if view.Winner == nil || *view.Winner != 0 {
		t.Fatalf("winner=%v", view.Winner)
	}
}

func driveStoreTimeoutPass(t *testing.T, store *Store, id BattleID) {
	t.Helper()
	for {
		b, err := store.Get(id)
		if err != nil {
			t.Fatal(err)
		}
		if b.Match.Finished {
			return
		}
		u0 := b.Match.Units[tactical.UnitIDPlayer0]
		u1 := b.Match.Units[tactical.UnitIDPlayer1]
		if _, _, err := store.SubmitTacticalCommand(id, 0, uint32(tactical.KindPass), tactical.UnitIDPlayer0, int32(u0.Pos.X), int32(u0.Pos.Y), 0); err != nil {
			t.Fatal(err)
		}
		if _, _, err := store.SubmitTacticalCommand(id, 1, uint32(tactical.KindPass), tactical.UnitIDPlayer1, int32(u1.Pos.X), int32(u1.Pos.Y), 0); err != nil {
			t.Fatal(err)
		}
		if _, _, _, _, _, err := store.StepLockstep(id, 1, StepLockstepOpts{}); err != nil {
			t.Fatal(err)
		}
		if b.Match.Frame > 66 {
			t.Fatal("timeout did not fire")
		}
	}
}
