package tactical

import (
	"errors"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/lockstep"
)

func TestCommandEncodeDecodeSkillPayload(t *testing.T) {
	in := Command{
		PlayerID: 0,
		Kind:     KindSkill,
		UnitID:   UnitIDPlayer0,
		To:       board.Coord{9, 9},
		SkillID:  combat.SkillStubStrike,
	}
	p := Encode(in)
	if len(p) != commandPayloadV2Len {
		t.Fatalf("want v2 payload len %d, got %d", commandPayloadV2Len, len(p))
	}
	out, err := Decode(p)
	if err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Fatalf("round trip mismatch: %+v vs %+v", in, out)
	}
	legacy, err := Decode(p[:commandPayloadV1Len])
	if err != nil {
		t.Fatal(err)
	}
	if legacy.SkillID != 0 {
		t.Fatalf("legacy decode should zero skill id, got %d", legacy.SkillID)
	}
}

func TestDefaultCardPoolContainsStubSkills(t *testing.T) {
	pool := combat.DefaultCardPool()
	if !pool.Contains(combat.SkillStubStrike) || !pool.Contains(combat.SkillStubSplash) {
		t.Fatalf("default pool missing stub skills: %+v", pool.Skills)
	}
	m := NewMatch(1)
	if !m.CardPool().Contains(combat.SkillStubStrike) {
		t.Fatal("match should inherit default card pool")
	}
}

func TestSubmitKindSkillUnknownRejected(t *testing.T) {
	m := NewMatch(2)
	err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindSkill,
		UnitID:   UnitIDPlayer0,
		To:       board.Coord{3, 8},
		SkillID:  99,
	})
	var se SkillError
	if !errors.As(err, &se) || se.Code != CodeSkillUnknown {
		t.Fatalf("want SKILL_UNKNOWN, got %v", err)
	}
}

func TestSubmitKindSkillStrikeAcceptedAdjacentEnemy(t *testing.T) {
	m := NewMatch(3)
	attacker := m.Units[UnitIDPlayer0]
	neighbor := board.Coord{X: attacker.Pos.X + 1, Y: attacker.Pos.Y}
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	m.Units[UnitIDPlayer1].Pos = neighbor
	if !m.Board.SetUnit(neighbor, UnitIDPlayer1) {
		t.Fatal("place defender")
	}
	err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindSkill,
		UnitID:   UnitIDPlayer0,
		To:       neighbor,
		SkillID:  combat.SkillStubStrike,
	})
	if err != nil {
		t.Fatalf("submit skill strike: %v", err)
	}
}

func TestSubmitKindSkillStrikeLockstepIntegration(t *testing.T) {
	m := NewMatch(4)
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
	if err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindSkill,
		UnitID:   UnitIDPlayer0,
		To:       neighbor,
		SkillID:  combat.SkillStubStrike,
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < int(lockstep.CommandDelayFrames)+1; i++ {
		m.StepLockstep()
	}
	if defender.Stats.HP.Raw() >= before {
		t.Fatalf("expected damage after skill strike lockstep, hp %d -> %d", before, defender.Stats.HP.Raw())
	}
}

func TestSubmitKindSkillStrikeViewSnapshotShowsCastAndDamage(t *testing.T) {
	m := NewMatch(6)
	attacker := m.Units[UnitIDPlayer0]
	neighbor := board.Coord{X: attacker.Pos.X + 1, Y: attacker.Pos.Y}
	m.Board.ClearUnit(m.Units[UnitIDPlayer1].Pos)
	defender := m.Units[UnitIDPlayer1]
	defender.Pos = neighbor
	m.Units[UnitIDPlayer1] = defender
	if !m.Board.SetUnit(neighbor, UnitIDPlayer1) {
		t.Fatal("place defender")
	}
	beforeHP := defender.Stats.HP.Raw()
	if err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindSkill,
		UnitID:   UnitIDPlayer0,
		To:       neighbor,
		SkillID:  combat.SkillStubStrike,
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < int(lockstep.CommandDelayFrames)+1; i++ {
		m.StepLockstep()
	}
	snap := MatchToViewSnapshot(m, 10000)
	if snap.LastSkillCast == nil {
		t.Fatal("expected lastSkillCast in view snapshot after skill resolve")
	}
	if snap.LastSkillCast.SkillID != uint16(combat.SkillStubStrike) {
		t.Fatalf("skill id=%d", snap.LastSkillCast.SkillID)
	}
	if snap.LastSkillCast.TargetX != neighbor.X || snap.LastSkillCast.TargetY != neighbor.Y {
		t.Fatalf("target=%d,%d want %d,%d", snap.LastSkillCast.TargetX, snap.LastSkillCast.TargetY, neighbor.X, neighbor.Y)
	}
	var enemyHP int64
	for _, u := range snap.Units {
		if u.ID == UnitIDPlayer1 {
			enemyHP = u.HP
		}
	}
	if enemyHP >= beforeHP {
		t.Fatalf("snapshot hp should reflect damage: before=%d after=%d", beforeHP, enemyHP)
	}
}

func TestSubmitKindSkillSplashLockstepIntegration(t *testing.T) {
	m := NewMatch(5)
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
	if err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindSkill,
		UnitID:   UnitIDPlayer0,
		To:       neighbor,
		SkillID:  combat.SkillStubSplash,
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < int(lockstep.CommandDelayFrames)+1; i++ {
		m.StepLockstep()
	}
	if defender.Stats.HP.Raw() >= before {
		t.Fatalf("expected splash damage, hp %d -> %d", before, defender.Stats.HP.Raw())
	}
}
