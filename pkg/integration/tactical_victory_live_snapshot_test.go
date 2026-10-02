package integration

import (
	"encoding/json"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
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

func TestPhase2OccupyViewSnapshotIntegration(t *testing.T) {
	m := tactical.NewMatchWithControlPoints(77, []tactical.ControlPoint{
		{Pos: board.Coord{2, 8}, HoldFrames: 2},
	})
	for i := 0; i < 2 && !m.Finished; i++ {
		m.StepLockstep()
	}
	raw, err := tactical.MarshalViewSnapshotJSON(tactical.MatchToViewSnapshot(m, 10000))
	if err != nil {
		t.Fatal(err)
	}
	var snap tactical.ViewSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if snap.EndReason != "occupy" {
		t.Fatalf("endReason=%q", snap.EndReason)
	}
	if snap.Winner == nil || *snap.Winner != 0 {
		t.Fatalf("winner=%v", snap.Winner)
	}
}

func TestPhase2TimeoutViewSnapshotIntegration(t *testing.T) {
	m := tactical.NewMatch(78)
	m.Units[tactical.UnitIDPlayer0].Stats.HP = fixed.FromInt(50)
	m.Units[tactical.UnitIDPlayer1].Stats.HP = fixed.FromInt(30)
	for !m.Finished {
		u0 := m.Units[tactical.UnitIDPlayer0]
		u1 := m.Units[tactical.UnitIDPlayer1]
		_ = m.Submit(tactical.Command{PlayerID: 0, Kind: tactical.KindPass, UnitID: tactical.UnitIDPlayer0, To: u0.Pos})
		_ = m.Submit(tactical.Command{PlayerID: 1, Kind: tactical.KindPass, UnitID: tactical.UnitIDPlayer1, To: u1.Pos})
		m.StepLockstep()
		if m.Frame > 66 {
			t.Fatal("timeout did not fire")
		}
	}
	raw, err := tactical.MarshalViewSnapshotJSON(tactical.MatchToViewSnapshot(m, 10000))
	if err != nil {
		t.Fatal(err)
	}
	var snap tactical.ViewSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if snap.EndReason != "timeout" {
		t.Fatalf("endReason=%q", snap.EndReason)
	}
	if snap.Winner == nil || *snap.Winner != 0 {
		t.Fatalf("winner=%v", snap.Winner)
	}
}
