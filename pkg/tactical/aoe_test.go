package tactical

import (
	"errors"
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

func TestApplyAoEStrikeDamagesAdjacentEnemy(t *testing.T) {
	m := NewMatch(2)
	attacker := m.Units[UnitIDPlayer0]
	center := attacker.Pos
	neighbor := board.Coord{X: center.X + 1, Y: center.Y}
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	defender := m.Units[UnitIDPlayer1]
	defender.Pos = neighbor
	m.Units[UnitIDPlayer1] = defender
	if !m.Board.SetUnit(neighbor, UnitIDPlayer1) {
		t.Fatal("place defender")
	}
	before := defender.Stats.HP.Raw()
	if err := m.ApplyAoEStrike(UnitIDPlayer0, center, combat.DefaultAoERadius); err != nil {
		t.Fatal(err)
	}
	if defender.Stats.HP.Raw() >= before {
		t.Fatalf("expected hp loss, before=%d after=%d", before, defender.Stats.HP.Raw())
	}
}

func TestApplyAoEStrikeNoEnemyRejected(t *testing.T) {
	m := NewMatch(3)
	center := m.Units[UnitIDPlayer0].Pos
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	err := m.ApplyAoEStrike(UnitIDPlayer0, center, combat.DefaultAoERadius)
	var ae AoEError
	if !errors.As(err, &ae) || ae.Code != CodeAoENoTargets {
		t.Fatalf("want AOE_NO_TARGETS, got %v", err)
	}
}

func TestSubmitKindAoEOutOfBoundsRejected(t *testing.T) {
	m := NewMatch(4)
	err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAoE,
		UnitID:   UnitIDPlayer0,
		To:       board.Coord{X: -1, Y: 0},
	})
	var ae AoEError
	if !errors.As(err, &ae) || ae.Code != CodeAoEOutOfBounds {
		t.Fatalf("want AOE_OUT_OF_BOUNDS, got %v", err)
	}
}

func TestSubmitKindAoEAcceptedWithNeighborEnemy(t *testing.T) {
	m := NewMatch(5)
	center := m.Units[UnitIDPlayer0].Pos
	neighbor := board.Coord{X: center.X + 1, Y: center.Y}
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	m.Units[UnitIDPlayer1].Pos = neighbor
	if !m.Board.SetUnit(neighbor, UnitIDPlayer1) {
		t.Fatal("place defender")
	}
	err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAoE,
		UnitID:   UnitIDPlayer0,
		To:       center,
	})
	if err != nil {
		t.Fatalf("submit aoe: %v", err)
	}
}
