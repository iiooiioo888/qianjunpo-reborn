package integration

import (
	"encoding/json"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

func TestPhase2VictoryViewSnapshotIntegration(t *testing.T) {
	m := tactical.NewMatch(88)
	def := m.Units[tactical.UnitIDPlayer1]
	def.Stats.HP = def.Stats.HP.Sub(def.Stats.HP)
	m.Board.ClearUnit(def.Pos)
	m.StepLockstep()

	raw, err := tactical.MarshalViewSnapshotJSON(tactical.MatchToViewSnapshot(m, 10000))
	if err != nil {
		t.Fatal(err)
	}
	var snap tactical.ViewSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if snap.EndReason != "wipeout" {
		t.Fatalf("endReason=%q", snap.EndReason)
	}
	if snap.Winner == nil || *snap.Winner != 0 {
		t.Fatalf("winner=%v", snap.Winner)
	}
}
