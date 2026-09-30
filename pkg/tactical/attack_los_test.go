package tactical

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/lockstep"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/validate"
)

// Symmetric to TestSubmitKindAoERangedLoSClearDamagesEnemy (#64): ranged KindAttack with clear LOS
// passes validateAttack and applies damage after lockstep.
func TestSubmitKindAttackRangedLoSClearDamagesEnemy(t *testing.T) {
	cfg, err := combat.DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	m := NewMatchWithConfig(19, cfg)
	archPos := board.Coord{X: 13, Y: 10}
	enemyPos := m.Units[UnitIDPlayer1].Pos
	if board.Chebyshev(archPos, enemyPos) != 3 {
		t.Fatalf("setup: want chebyshev 3, got %d", board.Chebyshev(archPos, enemyPos))
	}

	m.Board.ClearUnit(m.Units[UnitIDPlayer0].Pos)
	u := defaultUnit(cfg, UnitIDPlayer0, 0, combat.UnitArcher, archPos)
	m.Units[UnitIDPlayer0] = &u
	if !m.Board.SetUnit(archPos, UnitIDPlayer0) {
		t.Fatal("place archer")
	}

	enemyHPBefore := m.Units[UnitIDPlayer1].Stats.HP.Raw()
	if err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAttack,
		UnitID:   UnitIDPlayer0,
		To:       enemyPos,
	}); err != nil {
		t.Fatalf("submit ranged attack with clear LOS: %v", err)
	}
	for i := 0; i <= lockstep.CommandDelayFrames; i++ {
		m.StepLockstep()
	}
	if m.Units[UnitIDPlayer1].Stats.HP.Raw() >= enemyHPBefore {
		t.Fatalf("ranged attack: expected hp loss, before=%d after=%d", enemyHPBefore, m.Units[UnitIDPlayer1].Stats.HP.Raw())
	}
}

// Symmetric to TestSubmitKindAoELoSBlockedRejected (#64): third-party unit on trace rejects KindAttack.
func TestSubmitKindAttackLoSBlockedRejected(t *testing.T) {
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

	target := board.Coord{X: 5, Y: 2}
	blocker := board.Coord{X: 3, Y: 2}
	if !m.Board.SetUnit(blocker, 999) {
		t.Fatal("place blocker")
	}
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	defender := m.Units[UnitIDPlayer1]
	defender.Pos = target
	m.Units[UnitIDPlayer1] = defender
	if !m.Board.SetUnit(target, UnitIDPlayer1) {
		t.Fatal("place defender")
	}

	err = m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAttack,
		UnitID:   UnitIDPlayer0,
		To:       target,
	})
	if err == nil {
		t.Fatal("expected blocked line rejection")
	}
	var me validate.MoveError
	if !asMoveError(err, &me) || me.Code != validate.CodeBlocked {
		t.Fatalf("want CodeBlocked, got %v", err)
	}
}

// Symmetric to TestSubmitKindAttackLoSBlockedRejected: impassable terrain on trace rejects KindAttack.
func TestSubmitKindAttackLoSTerrainBlockedRejected(t *testing.T) {
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

	target := board.Coord{X: 5, Y: 2}
	blockCell := board.Coord{X: 3, Y: 2}
	m.Board.SetTerrain(blockCell, board.TerrainMountain)
	if m.Board.Get(blockCell).Passable {
		t.Fatal("setup: blocker terrain must be impassable")
	}
	if m.Board.GetUnit(blockCell) != 0 {
		t.Fatal("setup: blocker cell must be empty")
	}

	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	defender := m.Units[UnitIDPlayer1]
	defender.Pos = target
	m.Units[UnitIDPlayer1] = defender
	if !m.Board.SetUnit(target, UnitIDPlayer1) {
		t.Fatal("place defender")
	}

	err = m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAttack,
		UnitID:   UnitIDPlayer0,
		To:       target,
	})
	if err == nil {
		t.Fatal("expected terrain-blocked line rejection")
	}
	var me validate.MoveError
	if !asMoveError(err, &me) || me.Code != validate.CodeBlocked {
		t.Fatalf("want CodeBlocked, got %v", err)
	}
}

// Symmetric to TestSubmitKindAttackLoSBlockedRejected: same-line third-party unit at (3,2) would
// block ranged fire from (2,2)→(5,2), but adjacent melee KindAttack must not run AttackLineClear.
func TestSubmitKindAttackMeleeLoSNotCheckedUnitOnLineSubmitDamagesEnemy(t *testing.T) {
	cfg, err := combat.DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	m := NewMatchWithConfig(20, cfg)
	attPos := board.Coord{X: 4, Y: 2}
	target := board.Coord{X: 5, Y: 2}
	blocker := board.Coord{X: 3, Y: 2}

	m.Board.ClearUnit(m.Units[UnitIDPlayer0].Pos)
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	attacker := defaultUnit(cfg, UnitIDPlayer0, 0, combat.UnitInfantry, attPos)
	m.Units[UnitIDPlayer0] = &attacker
	if !m.Board.SetUnit(attPos, UnitIDPlayer0) {
		t.Fatal("place infantry")
	}
	if attacker.Stats.Range != 1 {
		t.Fatalf("setup: want melee range 1, got %d", attacker.Stats.Range)
	}
	if !m.Board.SetUnit(blocker, 999) {
		t.Fatal("place line blocker")
	}
	defender := defaultUnit(cfg, UnitIDPlayer1, 1, combat.UnitCavalry, target)
	m.Units[UnitIDPlayer1] = &defender
	if !m.Board.SetUnit(target, UnitIDPlayer1) {
		t.Fatal("place defender")
	}
	if board.Chebyshev(attPos, target) != 1 {
		t.Fatalf("setup: want chebyshev 1, got %d", board.Chebyshev(attPos, target))
	}

	enemyHPBefore := defender.Stats.HP.Raw()
	if err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAttack,
		UnitID:   UnitIDPlayer0,
		To:       target,
	}); err != nil {
		t.Fatalf("melee attack must ignore off-segment line blockers: %v", err)
	}
	for i := 0; i <= lockstep.CommandDelayFrames; i++ {
		m.StepLockstep()
	}
	if m.Units[UnitIDPlayer1].Stats.HP.Raw() >= enemyHPBefore {
		t.Fatalf("melee attack: expected hp loss, before=%d after=%d", enemyHPBefore, m.Units[UnitIDPlayer1].Stats.HP.Raw())
	}
}

// Symmetric to TestSubmitKindAttackLoSTerrainBlockedRejected: impassable terrain on the extended
// line must not reject adjacent melee KindAttack (RangedAttackNeedsLine false).
func TestSubmitKindAttackMeleeLoSNotCheckedTerrainOnLineSubmitDamagesEnemy(t *testing.T) {
	cfg, err := combat.DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	m := NewMatchWithConfig(21, cfg)
	attPos := board.Coord{X: 4, Y: 2}
	target := board.Coord{X: 5, Y: 2}
	blockCell := board.Coord{X: 3, Y: 2}

	m.Board.ClearUnit(m.Units[UnitIDPlayer0].Pos)
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	attacker := defaultUnit(cfg, UnitIDPlayer0, 0, combat.UnitInfantry, attPos)
	m.Units[UnitIDPlayer0] = &attacker
	if !m.Board.SetUnit(attPos, UnitIDPlayer0) {
		t.Fatal("place infantry")
	}
	m.Board.SetTerrain(blockCell, board.TerrainMountain)
	if m.Board.Get(blockCell).Passable {
		t.Fatal("setup: blocker terrain must be impassable")
	}
	defender := defaultUnit(cfg, UnitIDPlayer1, 1, combat.UnitCavalry, target)
	m.Units[UnitIDPlayer1] = &defender
	if !m.Board.SetUnit(target, UnitIDPlayer1) {
		t.Fatal("place defender")
	}

	enemyHPBefore := defender.Stats.HP.Raw()
	if err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAttack,
		UnitID:   UnitIDPlayer0,
		To:       target,
	}); err != nil {
		t.Fatalf("melee attack must not apply terrain LOS to adjacent target: %v", err)
	}
	for i := 0; i <= lockstep.CommandDelayFrames; i++ {
		m.StepLockstep()
	}
	if m.Units[UnitIDPlayer1].Stats.HP.Raw() >= enemyHPBefore {
		t.Fatalf("melee attack: expected hp loss, before=%d after=%d", enemyHPBefore, m.Units[UnitIDPlayer1].Stats.HP.Raw())
	}
}
