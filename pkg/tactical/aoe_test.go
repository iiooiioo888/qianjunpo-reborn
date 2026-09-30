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
	neighbor := board.Coord{X: attacker.Pos.X + 1, Y: attacker.Pos.Y}
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	defender := m.Units[UnitIDPlayer1]
	defender.Pos = neighbor
	m.Units[UnitIDPlayer1] = defender
	if !m.Board.SetUnit(neighbor, UnitIDPlayer1) {
		t.Fatal("place defender")
	}
	before := defender.Stats.HP.Raw()
	if err := m.ApplyAoEStrike(UnitIDPlayer0, neighbor, combat.DefaultAoERadius); err != nil {
		t.Fatal(err)
	}
	if defender.Stats.HP.Raw() >= before {
		t.Fatalf("expected hp loss, before=%d after=%d", before, defender.Stats.HP.Raw())
	}
}

func TestApplyAoEStrikeNoEnemyRejected(t *testing.T) {
	m := NewMatch(3)
	attacker := m.Units[UnitIDPlayer0]
	center := board.Coord{X: attacker.Pos.X + 1, Y: attacker.Pos.Y}
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
	attacker := m.Units[UnitIDPlayer0]
	neighbor := board.Coord{X: attacker.Pos.X + 1, Y: attacker.Pos.Y}
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	m.Units[UnitIDPlayer1].Pos = neighbor
	if !m.Board.SetUnit(neighbor, UnitIDPlayer1) {
		t.Fatal("place defender")
	}
	err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAoE,
		UnitID:   UnitIDPlayer0,
		To:       neighbor,
	})
	if err != nil {
		t.Fatalf("submit aoe: %v", err)
	}
}

func TestApplyAoEStrikeOutOfRangeRejected(t *testing.T) {
	m := NewMatch(6)
	attacker := m.Units[UnitIDPlayer0]
	far := board.Coord{X: attacker.Pos.X + 3, Y: attacker.Pos.Y}
	err := m.ApplyAoEStrike(UnitIDPlayer0, far, combat.DefaultAoERadius)
	var ae AoEError
	if !errors.As(err, &ae) || ae.Code != CodeAoEOutOfRange {
		t.Fatalf("want AOE_OUT_OF_RANGE, got %v", err)
	}
}

func TestSubmitKindAoELoSBlockedRejected(t *testing.T) {
	cfg, err := combat.DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	m := NewMatchWithConfig(7, cfg)
	attacker := m.Units[UnitIDPlayer0]
	attacker.Stats.Type = combat.UnitArcher
	attacker.Stats.Range = 3
	attacker.Pos = board.Coord{X: 2, Y: 2}
	m.Units[UnitIDPlayer0] = attacker
	m.Board.ClearUnit(board.Coord{2, 8})
	if !m.Board.SetUnit(attacker.Pos, UnitIDPlayer0) {
		t.Fatal("place archer")
	}

	center := board.Coord{X: 5, Y: 2}
	blocker := board.Coord{X: 3, Y: 2}
	if !m.Board.SetUnit(blocker, 999) {
		t.Fatal("place blocker")
	}
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	defender := m.Units[UnitIDPlayer1]
	defender.Pos = center
	m.Units[UnitIDPlayer1] = defender
	if !m.Board.SetUnit(center, UnitIDPlayer1) {
		t.Fatal("place defender")
	}

	err = m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAoE,
		UnitID:   UnitIDPlayer0,
		To:       center,
	})
	var ae AoEError
	if !errors.As(err, &ae) || ae.Code != CodeAoELoSBlocked {
		t.Fatalf("want AOE_LOS_BLOCKED, got %v", err)
	}
}
