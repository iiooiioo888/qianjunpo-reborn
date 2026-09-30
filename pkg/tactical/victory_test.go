package tactical

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
)

func TestAnnihilationVictory(t *testing.T) {
	m := NewMatch(42)
	def := m.Units[UnitIDPlayer1]
	def.Stats.HP = fixed.Zero
	m.Board.ClearUnit(def.Pos)
	m.checkVictory()
	if !m.Finished || m.Winner != 0 || m.EndReason != EndAnnihilation {
		t.Fatalf("expected P0 annihilation win, got finished=%v winner=%d reason=%d", m.Finished, m.Winner, m.EndReason)
	}
}

func TestMutualWipeDraw(t *testing.T) {
	m := NewMatch(43)
	for _, u := range m.Units {
		u.Stats.HP = fixed.Zero
		m.Board.ClearUnit(u.Pos)
	}
	m.checkVictory()
	if !m.Finished || m.Winner != NoWinner || m.EndReason != EndMutualWipe {
		t.Fatalf("expected mutual wipe draw, got winner=%d reason=%d", m.Winner, m.EndReason)
	}
}

func TestTimeoutHigherHPWins(t *testing.T) {
	m := NewMatch(44)
	m.Units[UnitIDPlayer0].Stats.HP = fixed.FromInt(50)
	m.Units[UnitIDPlayer1].Stats.HP = fixed.FromInt(30)
	m.Frame = maxTurnFrames
	m.decideTimeoutWinner()
	if m.Winner != 0 || m.EndReason != EndTimeout {
		t.Fatalf("expected timeout win P0, got winner=%d reason=%d", m.Winner, m.EndReason)
	}
}

func TestTimeoutTieIsDraw(t *testing.T) {
	m := NewMatch(45)
	hp := fixed.FromInt(40)
	m.Units[UnitIDPlayer0].Stats.HP = hp
	m.Units[UnitIDPlayer1].Stats.HP = hp
	m.Frame = maxTurnFrames
	m.decideTimeoutWinner()
	if m.Winner != NoWinner || m.EndReason != EndTimeout {
		t.Fatalf("expected timeout draw, got winner=%d reason=%d", m.Winner, m.EndReason)
	}
}
