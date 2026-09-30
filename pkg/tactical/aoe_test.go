package tactical

import (
	"errors"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/lockstep"
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

func TestApplyAoEStrikeFriendlyCenterSkipsAllies(t *testing.T) {
	cfg, err := combat.DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	m := NewMatchWithConfig(8, cfg)
	const allyID uint32 = 102

	attPos := board.Coord{X: 5, Y: 8}
	allyPos := board.Coord{X: 6, Y: 8}
	enemyPos := board.Coord{X: 7, Y: 8}

	m.Board.ClearUnit(m.Units[UnitIDPlayer0].Pos)
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	attacker := defaultUnit(cfg, UnitIDPlayer0, 0, combat.UnitInfantry, attPos)
	m.Units[UnitIDPlayer0] = &attacker
	if !m.Board.SetUnit(attPos, UnitIDPlayer0) {
		t.Fatal("place attacker")
	}
	ally := defaultUnit(cfg, allyID, 0, combat.UnitInfantry, allyPos)
	m.Units[allyID] = &ally
	if !m.Board.SetUnit(allyPos, allyID) {
		t.Fatal("place ally")
	}
	defender := defaultUnit(cfg, UnitIDPlayer1, 1, combat.UnitCavalry, enemyPos)
	m.Units[UnitIDPlayer1] = &defender
	if !m.Board.SetUnit(enemyPos, UnitIDPlayer1) {
		t.Fatal("place defender")
	}

	allyHPBefore := m.Units[allyID].Stats.HP.Raw()
	enemyHPBefore := defender.Stats.HP.Raw()
	if err := m.ApplyAoEStrike(UnitIDPlayer0, allyPos, combat.DefaultAoERadius); err != nil {
		t.Fatal(err)
	}
	if m.Units[allyID].Stats.HP.Raw() != allyHPBefore {
		t.Fatalf("ally took splash damage: before=%d after=%d", allyHPBefore, m.Units[allyID].Stats.HP.Raw())
	}
	if m.Units[UnitIDPlayer1].Stats.HP.Raw() >= enemyHPBefore {
		t.Fatalf("expected enemy hp loss, before=%d after=%d", enemyHPBefore, m.Units[UnitIDPlayer1].Stats.HP.Raw())
	}
}

func TestSubmitKindAoEFriendlyCenterLockstepIntegration(t *testing.T) {
	cfg, err := combat.DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	m := NewMatchWithConfig(9, cfg)
	const allyID uint32 = 103

	attPos := board.Coord{X: 5, Y: 8}
	allyPos := board.Coord{X: 6, Y: 8}
	enemyPos := board.Coord{X: 7, Y: 8}

	m.Board.ClearUnit(m.Units[UnitIDPlayer0].Pos)
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	attacker := defaultUnit(cfg, UnitIDPlayer0, 0, combat.UnitInfantry, attPos)
	m.Units[UnitIDPlayer0] = &attacker
	if !m.Board.SetUnit(attPos, UnitIDPlayer0) {
		t.Fatal("place attacker")
	}
	ally := defaultUnit(cfg, allyID, 0, combat.UnitInfantry, allyPos)
	m.Units[allyID] = &ally
	if !m.Board.SetUnit(allyPos, allyID) {
		t.Fatal("place ally")
	}
	defender := defaultUnit(cfg, UnitIDPlayer1, 1, combat.UnitCavalry, enemyPos)
	m.Units[UnitIDPlayer1] = &defender
	if !m.Board.SetUnit(enemyPos, UnitIDPlayer1) {
		t.Fatal("place defender")
	}

	enemyHPBefore := defender.Stats.HP.Raw()
	if err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAoE,
		UnitID:   UnitIDPlayer0,
		To:       allyPos,
	}); err != nil {
		t.Fatalf("submit aoe on friendly center: %v", err)
	}
	for i := 0; i <= lockstep.CommandDelayFrames; i++ {
		m.StepLockstep()
	}
	if m.Units[UnitIDPlayer1].Stats.HP.Raw() >= enemyHPBefore {
		t.Fatalf("lockstep aoe: expected enemy hp loss, before=%d after=%d", enemyHPBefore, m.Units[UnitIDPlayer1].Stats.HP.Raw())
	}
}

func TestSubmitKindAoERangedLoSClearDamagesEnemy(t *testing.T) {
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
		Kind:     KindAoE,
		UnitID:   UnitIDPlayer0,
		To:       enemyPos,
	}); err != nil {
		t.Fatalf("submit ranged aoe with clear LOS: %v", err)
	}
	for i := 0; i <= lockstep.CommandDelayFrames; i++ {
		m.StepLockstep()
	}
	if m.Units[UnitIDPlayer1].Stats.HP.Raw() >= enemyHPBefore {
		t.Fatalf("ranged aoe: expected hp loss, before=%d after=%d", enemyHPBefore, m.Units[UnitIDPlayer1].Stats.HP.Raw())
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

// Symmetric to TestSubmitKindAttackLoSTerrainBlockedRejected (#69): impassable terrain on trace rejects KindAoE.
func TestSubmitKindAoELoSTerrainBlockedRejected(t *testing.T) {
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
	if err == nil {
		t.Fatal("expected terrain-blocked line rejection")
	}
	var ae AoEError
	if !errors.As(err, &ae) || ae.Code != CodeAoELoSBlocked {
		t.Fatalf("want AOE_LOS_BLOCKED, got %v", err)
	}
}

// Symmetric to TestSubmitKindAoELoSBlockedRejected: line blocker at (3,2) would reject ranged AoE from
// (2,2)→(5,2), but adjacent melee KindAoE must skip AttackLineClear and damage after lockstep.
func TestSubmitKindAoEMeleeLoSNotCheckedUnitOnLineSubmitDamagesEnemy(t *testing.T) {
	cfg, err := combat.DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	m := NewMatchWithConfig(22, cfg)
	attPos := board.Coord{X: 4, Y: 2}
	center := board.Coord{X: 5, Y: 2}
	blocker := board.Coord{X: 3, Y: 2}

	m.Board.ClearUnit(m.Units[UnitIDPlayer0].Pos)
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	attacker := defaultUnit(cfg, UnitIDPlayer0, 0, combat.UnitInfantry, attPos)
	m.Units[UnitIDPlayer0] = &attacker
	if !m.Board.SetUnit(attPos, UnitIDPlayer0) {
		t.Fatal("place infantry")
	}
	if !m.Board.SetUnit(blocker, 999) {
		t.Fatal("place line blocker")
	}
	defender := defaultUnit(cfg, UnitIDPlayer1, 1, combat.UnitCavalry, center)
	m.Units[UnitIDPlayer1] = &defender
	if !m.Board.SetUnit(center, UnitIDPlayer1) {
		t.Fatal("place defender")
	}

	enemyHPBefore := defender.Stats.HP.Raw()
	if err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAoE,
		UnitID:   UnitIDPlayer0,
		To:       center,
	}); err != nil {
		t.Fatalf("melee aoe must ignore off-segment line blockers: %v", err)
	}
	for i := 0; i <= lockstep.CommandDelayFrames; i++ {
		m.StepLockstep()
	}
	if m.Units[UnitIDPlayer1].Stats.HP.Raw() >= enemyHPBefore {
		t.Fatalf("melee aoe: expected hp loss, before=%d after=%d", enemyHPBefore, m.Units[UnitIDPlayer1].Stats.HP.Raw())
	}
}

// Symmetric to TestSubmitKindAoELoSTerrainBlockedRejected: impassable terrain on extended line must
// not reject adjacent melee KindAoE.
func TestSubmitKindAoEMeleeLoSNotCheckedTerrainOnLineSubmitDamagesEnemy(t *testing.T) {
	cfg, err := combat.DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	m := NewMatchWithConfig(23, cfg)
	attPos := board.Coord{X: 4, Y: 2}
	center := board.Coord{X: 5, Y: 2}
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
	defender := defaultUnit(cfg, UnitIDPlayer1, 1, combat.UnitCavalry, center)
	m.Units[UnitIDPlayer1] = &defender
	if !m.Board.SetUnit(center, UnitIDPlayer1) {
		t.Fatal("place defender")
	}

	enemyHPBefore := defender.Stats.HP.Raw()
	if err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindAoE,
		UnitID:   UnitIDPlayer0,
		To:       center,
	}); err != nil {
		t.Fatalf("melee aoe must not apply terrain LOS to adjacent center: %v", err)
	}
	for i := 0; i <= lockstep.CommandDelayFrames; i++ {
		m.StepLockstep()
	}
	if m.Units[UnitIDPlayer1].Stats.HP.Raw() >= enemyHPBefore {
		t.Fatalf("melee aoe: expected hp loss, before=%d after=%d", enemyHPBefore, m.Units[UnitIDPlayer1].Stats.HP.Raw())
	}
}
