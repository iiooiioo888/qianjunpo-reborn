package tactical

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
)

func TestMatchCollectAoETargets_stub(t *testing.T) {
	m := NewMatch(1)
	center := m.Units[UnitIDPlayer0].Pos
	neighbor := board.Coord{X: center.X + 1, Y: center.Y}
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	m.Units[UnitIDPlayer1].Pos = neighbor
	if !m.Board.SetUnit(neighbor, UnitIDPlayer1) {
		t.Fatal("set player1 adjacent")
	}

	ids := m.CollectAoETargets(center, combat.DefaultAoERadius, UnitIDPlayer0)
	if len(ids) != 1 || ids[0] != UnitIDPlayer1 {
		t.Fatalf("want adjacent defender in splash, got %v", ids)
	}
}
