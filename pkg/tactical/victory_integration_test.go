package tactical

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
)

// Lockstep integration for capture / annihilation / timeout (separate from attack LOS tests in attack_los_test.go).

func TestCapturePointHoldLockstepIntegration(t *testing.T) {
	// P0 spawns on (2,8); sole occupancy ticks advance each StepLockstep.
	m := NewMatchWithControlPoints(100, []ControlPoint{
		{Pos: board.Coord{2, 8}, HoldFrames: 3},
	})
	for i := 0; i < 3 && !m.Finished; i++ {
		m.StepLockstep()
	}
	if !m.Finished || m.Winner != 0 || m.EndReason != EndCapture {
		t.Fatalf("expected P0 capture win, finished=%v winner=%d reason=%d frame=%d", m.Finished, m.Winner, m.EndReason, m.Frame)
	}
}

func TestAnnihilationLockstepIntegration(t *testing.T) {
	m := NewMatch(101)
	enemy := m.Units[UnitIDPlayer1]
	enemy.Stats.HP = enemy.Stats.HP.Sub(enemy.Stats.HP)
	m.Board.ClearUnit(enemy.Pos)
	m.StepLockstep()
	if !m.Finished || m.Winner != 0 || m.EndReason != EndAnnihilation {
		t.Fatalf("expected annihilation, finished=%v winner=%d reason=%d", m.Finished, m.Winner, m.EndReason)
	}
}

func TestTimeoutPassLockstepIntegration(t *testing.T) {
	m := NewMatch(102)
	for !m.Finished {
		_ = m.Submit(Command{PlayerID: 0, Kind: KindPass, UnitID: UnitIDPlayer0, To: m.Units[UnitIDPlayer0].Pos})
		_ = m.Submit(Command{PlayerID: 1, Kind: KindPass, UnitID: UnitIDPlayer1, To: m.Units[UnitIDPlayer1].Pos})
		m.StepLockstep()
		if m.Frame > maxTurnFrames+2 {
			t.Fatal("timeout did not fire")
		}
	}
	if !m.Finished || m.EndReason != EndTimeout {
		t.Fatalf("expected timeout finish, finished=%v reason=%d winner=%d", m.Finished, m.EndReason, m.Winner)
	}
}
