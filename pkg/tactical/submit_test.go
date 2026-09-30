package tactical

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/validate"
)

func TestSubmitWrongOwnerRejected(t *testing.T) {
	m := NewMatch(10)
	err := m.Submit(Command{
		PlayerID: 1,
		Kind:     KindMove,
		UnitID:   UnitIDPlayer0,
		To:       board.Coord{3, 8},
	})
	if err == nil {
		t.Fatal("expected wrong owner error")
	}
	var me validate.MoveError
	if !asMoveError(err, &me) || me.Code != validate.CodeWrongStart {
		t.Fatalf("expected CodeWrongStart, got %v", err)
	}
}

func TestSubmitAttackNoTargetRejected(t *testing.T) {
	m := NewMatch(11)
	err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAttack,
		UnitID:   UnitIDPlayer0,
		To:       board.Coord{3, 8},
	})
	if err == nil {
		t.Fatal("expected no target error")
	}
}

func TestSubmitAttackFriendlyRejected(t *testing.T) {
	m := NewMatch(12)
	err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAttack,
		UnitID:   UnitIDPlayer0,
		To:       m.Units[UnitIDPlayer0].Pos,
	})
	if err == nil {
		t.Fatal("expected friendly fire rejection")
	}
}

func TestSubmitAfterFinishedRejected(t *testing.T) {
	m := NewMatch(13)
	m.Finished = true
	err := m.Submit(Command{PlayerID: 0, Kind: KindPass, UnitID: UnitIDPlayer0, To: m.Units[UnitIDPlayer0].Pos})
	if err == nil {
		t.Fatal("expected finished match rejection")
	}
}

func TestSubmitInvalidPlayerRejected(t *testing.T) {
	m := NewMatch(14)
	err := m.Submit(Command{
		PlayerID: 9,
		Kind:     KindPass,
		UnitID:   UnitIDPlayer0,
		To:       m.Units[UnitIDPlayer0].Pos,
	})
	if err == nil {
		t.Fatal("expected invalid player error")
	}
}

func TestSubmitAttackOutOfRangeRejected(t *testing.T) {
	m := NewMatch(15)
	// Enemy is far away; empty cell at distance 2 is not a valid melee attack.
	err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAttack,
		UnitID:   UnitIDPlayer0,
		To:       board.Coord{4, 8},
	})
	if err == nil {
		t.Fatal("expected out of range / no target")
	}
}

func TestArcherRangeAllowsDistanceTwo(t *testing.T) {
	cfg, err := combat.DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	m := NewMatchWithConfig(16, cfg)
	// Replace P0 with archer in open space near P1.
	archPos := board.Coord{14, 10}
	m.Board.ClearUnit(m.Units[UnitIDPlayer0].Pos)
	u := defaultUnit(cfg, UnitIDPlayer0, 0, combat.UnitArcher, archPos)
	m.Units[UnitIDPlayer0] = &u
	m.Board.SetUnit(archPos, UnitIDPlayer0)
	enemyPos := m.Units[UnitIDPlayer1].Pos
	if board.Chebyshev(archPos, enemyPos) != 2 {
		t.Fatalf("test setup: expected chebyshev 2, got %d", board.Chebyshev(archPos, enemyPos))
	}
	err = m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAttack,
		UnitID:   UnitIDPlayer0,
		To:       enemyPos,
	})
	if err != nil {
		t.Fatalf("archer should attack at range 2: %v", err)
	}
}
