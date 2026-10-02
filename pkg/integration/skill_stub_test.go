package integration

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/lockstep"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

// Cross-package smoke: card pool + KindSkill enter Submit/lockstep without touching counters.
func TestPhase2SkillStubIntegration(t *testing.T) {
	pool := combat.DefaultCardPool()
	if pool.ID != combat.DefaultCardPoolID {
		t.Fatalf("unexpected pool id %d", pool.ID)
	}
	m := tactical.NewMatch(42)
	if m.CardPool().ID != pool.ID {
		t.Fatal("match card pool not wired")
	}
	attacker := m.Units[tactical.UnitIDPlayer0]
	target := board.Coord{X: attacker.Pos.X + 1, Y: attacker.Pos.Y}
	m.Board.ClearUnit(m.Units[tactical.UnitIDPlayer1].Pos)
	def := m.Units[tactical.UnitIDPlayer1]
	def.Pos = target
	m.Units[tactical.UnitIDPlayer1] = def
	if !m.Board.SetUnit(target, tactical.UnitIDPlayer1) {
		t.Fatal("place enemy")
	}
	if err := m.Submit(tactical.Command{
		PlayerID: 0,
		Kind:     tactical.KindSkill,
		UnitID:   tactical.UnitIDPlayer0,
		To:       target,
		SkillID:  combat.SkillStubStrike,
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < int(lockstep.CommandDelayFrames)+1; i++ {
		m.StepLockstep()
	}
	if def.Stats.HP.Raw() <= 0 {
		return
	}
	if def.Stats.HP.Raw() >= def.Stats.BaseHP.Raw() {
		t.Fatal("skill stub strike should reduce enemy hp")
	}
}
