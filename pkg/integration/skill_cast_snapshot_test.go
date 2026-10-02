package integration

import (
	"encoding/json"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/lockstep"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

// End-to-end: KindSkill via Submit → lockstep → ViewSnapshot JSON shows cast metadata + HP drop.
func TestPhase2SkillCastViewSnapshotIntegration(t *testing.T) {
	m := tactical.NewMatch(77)
	attacker := m.Units[tactical.UnitIDPlayer0]
	target := board.Coord{X: attacker.Pos.X + 1, Y: attacker.Pos.Y}
	m.Board.ClearUnit(m.Units[tactical.UnitIDPlayer1].Pos)
	def := m.Units[tactical.UnitIDPlayer1]
	def.Pos = target
	m.Units[tactical.UnitIDPlayer1] = def
	if !m.Board.SetUnit(target, tactical.UnitIDPlayer1) {
		t.Fatal("place enemy")
	}
	before := def.Stats.HP.Raw()
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
	raw, err := tactical.MarshalViewSnapshotJSON(tactical.MatchToViewSnapshot(m, 10000))
	if err != nil {
		t.Fatal(err)
	}
	var snap tactical.ViewSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if snap.LastSkillCast == nil || snap.LastSkillCast.SkillID != uint16(combat.SkillStubStrike) {
		t.Fatalf("lastSkillCast=%v", snap.LastSkillCast)
	}
	var after int64
	for _, u := range snap.Units {
		if u.ID == tactical.UnitIDPlayer1 {
			after = u.HP
		}
	}
	if after >= before {
		t.Fatalf("expected damage in snapshot units, hp %d -> %d", before, after)
	}
}
