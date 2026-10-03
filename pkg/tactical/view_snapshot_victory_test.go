package tactical

import (
	"encoding/json"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
)

func TestViewSnapshotInProgressHasNoWinner(t *testing.T) {
	m := NewMatch(1)
	snap := MatchToViewSnapshot(m, 10000)
	if snap.Winner != nil {
		t.Fatalf("expected nil winner in progress, got %v", snap.Winner)
	}
	if snap.EndReason != "none" {
		t.Fatalf("endReason=%q", snap.EndReason)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	if !jsonContainsNullWinner(raw) {
		t.Fatalf("expected winner:null in JSON, got %s", raw)
	}
}

func jsonContainsNullWinner(raw []byte) bool {
	var generic map[string]interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		return false
	}
	w, ok := generic["winner"]
	return ok && w == nil
}

func TestViewSnapshotAnnihilationExportsWinnerAndReason(t *testing.T) {
	m := NewMatch(101)
	enemy := m.Units[UnitIDPlayer1]
	enemy.Stats.HP = enemy.Stats.HP.Sub(enemy.Stats.HP)
	m.Board.ClearUnit(enemy.Pos)
	m.StepLockstep()
	snap := MatchToViewSnapshot(m, 10000)
	if snap.Winner == nil || *snap.Winner != 0 {
		t.Fatalf("winner=%v", snap.Winner)
	}
	if snap.EndReason != "wipeout" {
		t.Fatalf("endReason=%q", snap.EndReason)
	}
}

func TestViewSnapshotCaptureExportsOccupy(t *testing.T) {
	m := NewMatchWithControlPoints(100, []ControlPoint{
		{Pos: board.Coord{2, 8}, HoldFrames: 2},
	})
	for i := 0; i < 2 && !m.Finished; i++ {
		m.StepLockstep()
	}
	if !m.Finished {
		t.Fatal("expected occupy finish after hold frames")
	}
	snap := MatchToViewSnapshot(m, 10000)
	if snap.EndReason != "occupy" || snap.Winner == nil || *snap.Winner != 0 {
		t.Fatalf("snap winner=%v reason=%q finished=%v", snap.Winner, snap.EndReason, m.Finished)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var generic map[string]interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatal(err)
	}
	if generic["endReason"] != "occupy" {
		t.Fatalf("json endReason=%v", generic["endReason"])
	}
}

func TestViewSnapshotTimeoutExportsWinnerAndReason(t *testing.T) {
	m := NewMatch(102)
	m.Units[UnitIDPlayer0].Stats.HP = fixed.FromInt(50)
	m.Units[UnitIDPlayer1].Stats.HP = fixed.FromInt(30)
	driveMatchToTimeoutPass(t, m)
	snap := MatchToViewSnapshot(m, 10000)
	if snap.EndReason != "timeout" {
		t.Fatalf("endReason=%q", snap.EndReason)
	}
	if snap.Winner == nil || *snap.Winner != 0 {
		t.Fatalf("winner=%v", snap.Winner)
	}
}

func TestViewSnapshotTimeoutTieExportsDrawSentinel(t *testing.T) {
	m := NewMatch(103)
	hp := fixed.FromInt(40)
	m.Units[UnitIDPlayer0].Stats.HP = hp
	m.Units[UnitIDPlayer1].Stats.HP = hp
	driveMatchToTimeoutPass(t, m)
	snap := MatchToViewSnapshot(m, 10000)
	if snap.EndReason != "timeout" {
		t.Fatalf("endReason=%q", snap.EndReason)
	}
	if snap.Winner == nil || *snap.Winner != NoWinner {
		t.Fatalf("expected draw sentinel %d, winner=%v", NoWinner, snap.Winner)
	}
}

func driveMatchToTimeoutPass(t *testing.T, m *Match) {
	t.Helper()
	for !m.Finished {
		u0 := m.Units[UnitIDPlayer0]
		u1 := m.Units[UnitIDPlayer1]
		_ = m.Submit(Command{PlayerID: 0, Kind: KindPass, UnitID: UnitIDPlayer0, To: u0.Pos})
		_ = m.Submit(Command{PlayerID: 1, Kind: KindPass, UnitID: UnitIDPlayer1, To: u1.Pos})
		m.StepLockstep()
		if m.Frame > MaxTurnFrames+2 {
			t.Fatal("timeout did not fire")
		}
	}
}
