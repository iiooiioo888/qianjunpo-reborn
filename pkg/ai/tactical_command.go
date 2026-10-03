package ai

import (
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

// TacticalCommandJSON mirrors Janus POST /v1/tactical/command (snake_case).
// Auto-battle pollers can POST this object as-is after filling battle_id when omitted.
type TacticalCommandJSON struct {
	BattleID  string `json:"battle_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	PlayerID  uint32 `json:"player_id"`
	Kind      uint32 `json:"kind"`
	UnitID    uint32 `json:"unit_id"`
	ToX       int32  `json:"to_x"`
	ToY       int32  `json:"to_y"`
	SkillID   uint32 `json:"skill_id,omitempty"`
}

// TacticalCommandFromSuggestion maps a consumable hint into Janus tactical command fields.
func TacticalCommandFromSuggestion(battleID string, sug InferSuggestion) TacticalCommandJSON {
	kind := uint32(tactical.KindMove)
	if sug.Kind == SuggestionKindSkill {
		kind = uint32(tactical.KindSkill)
	}
	out := TacticalCommandJSON{
		BattleID: battleID,
		PlayerID: 0,
		Kind:     kind,
		UnitID:   sug.UnitID,
		ToX:      int32(sug.ToX),
		ToY:      int32(sug.ToY),
	}
	if kind == uint32(tactical.KindSkill) && sug.SkillID != 0 {
		out.SkillID = sug.SkillID
	}
	return out
}

// AttachTacticalCommand sets resp.Command from resp.Suggestion when present.
func AttachTacticalCommand(resp *InferHTTPResponse, battleID string) {
	if resp == nil || resp.Suggestion == nil {
		return
	}
	cmd := TacticalCommandFromSuggestion(battleID, *resp.Suggestion)
	resp.Command = &cmd
}
