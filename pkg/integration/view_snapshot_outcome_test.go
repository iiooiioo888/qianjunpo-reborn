package integration

import (
	"encoding/json"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/internal/roma"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

// Roma TacticalViewSnapshotJSON passthrough includes finished/winner/endReason after annihilation.
func TestPhase2ViewSnapshotOutcomeRomaIntegration(t *testing.T) {
	store := roma.NewStore(nil)
	b, err := store.Join("default", 0)
	if err != nil {
		t.Fatal(err)
	}
	def := b.Match.Units[tactical.UnitIDPlayer1]
	def.Stats.HP = fixed.Zero
	b.Match.Board.ClearUnit(def.Pos)
	_, _, finished, _, err := store.StepLockstep(b.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !finished || !b.Match.Finished {
		t.Fatal("expected finished match after annihilation step")
	}

	raw, _, _, err := store.TacticalViewSnapshotJSON(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	var snap tactical.ViewSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if !snap.Finished || snap.Winner == nil || *snap.Winner != 0 || snap.EndReason != tactical.EndAnnihilation {
		t.Fatalf("outcome: finished=%v winner=%v reason=%s", snap.Finished, snap.Winner, snap.EndReason)
	}
}
