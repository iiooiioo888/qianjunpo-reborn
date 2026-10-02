package roma

import (
	"encoding/json"
	"testing"

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
	_, _, finished, winner, err := store.StepLockstep(b.ID, 1)
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
