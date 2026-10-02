package roma

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

func TestProtoToCommandKindSkillMapsSkillID(t *testing.T) {
	cmd, err := protoToCommand(0, uint32(tactical.KindSkill), tactical.UnitIDPlayer0, 9, 9, uint32(combat.SkillStubStrike))
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Kind != tactical.KindSkill || cmd.SkillID != combat.SkillStubStrike {
		t.Fatalf("got %+v", cmd)
	}
}

func TestProtoToCommandKindSkillRequiresSkillID(t *testing.T) {
	_, err := protoToCommand(0, uint32(tactical.KindSkill), tactical.UnitIDPlayer0, 9, 9, 0)
	if err == nil || err.Error() != "roma: kind skill requires skill_id" {
		t.Fatalf("want skill_id required, got %v", err)
	}
}

func TestProtoToCommandUnknownSkillRejectedByMatch(t *testing.T) {
	store := NewStore(nil)
	b, err := store.Join("default", 0)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = store.SubmitTacticalCommand(b.ID, 0, uint32(tactical.KindSkill), tactical.UnitIDPlayer0, 9, 9, 99)
	if err == nil {
		t.Fatal("expected unknown skill rejection from Match.Submit")
	}
}

func TestProtoToCommandMoveIgnoresSkillIDWire(t *testing.T) {
	cmd, err := protoToCommand(0, uint32(tactical.KindMove), tactical.UnitIDPlayer0, 5, 8, uint32(combat.SkillStubStrike))
	if err != nil {
		t.Fatal(err)
	}
	if cmd.SkillID != 0 {
		t.Fatalf("move should not carry skill id, got %d", cmd.SkillID)
	}
}
