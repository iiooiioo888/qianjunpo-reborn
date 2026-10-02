package tactical

import (
	"fmt"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
)

// Skill error codes for submit-time validation (stub wiring).
const (
	CodeSkillUnknown    = "SKILL_UNKNOWN"
	CodeSkillNotInPool  = "SKILL_NOT_IN_POOL"
	CodeSkillNoBehavior = "SKILL_NO_BEHAVIOR"
)

// SkillError is returned when a KindSkill command fails validation.
type SkillError struct {
	Code    string
	Message string
}

func (e SkillError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

func (m *Match) validateSkill(cmd Command, caster *Unit) error {
	if cmd.SkillID == 0 {
		return SkillError{Code: CodeSkillUnknown, Message: "missing skill id"}
	}
	def, ok := combat.LookupSkill(cmd.SkillID)
	if !ok {
		return SkillError{Code: CodeSkillUnknown, Message: fmt.Sprintf("unknown skill %d", cmd.SkillID)}
	}
	if !m.cardPool.Contains(cmd.SkillID) {
		return SkillError{Code: CodeSkillNotInPool, Message: "skill not in match card pool"}
	}
	switch def.Behavior {
	case combat.SkillBehaviorStrike:
		return m.validateAttack(cmd, caster)
	case combat.SkillBehaviorSplash:
		return m.validateAoE(cmd, caster)
	default:
		return SkillError{Code: CodeSkillNoBehavior, Message: "skill behavior not implemented"}
	}
}

func (m *Match) applySkill(cmd Command, caster *Unit) {
	def, ok := combat.LookupSkill(cmd.SkillID)
	if !ok || !m.cardPool.Contains(cmd.SkillID) {
		return
	}
	switch def.Behavior {
	case combat.SkillBehaviorStrike:
		if err := m.validateAttack(cmd, caster); err != nil {
			return
		}
		targetID := m.Board.GetUnit(cmd.To)
		defender := m.Units[targetID]
		if defender == nil {
			return
		}
		m.applyStrike(caster, defender)
	case combat.SkillBehaviorSplash:
		if err := m.validateAoE(cmd, caster); err != nil {
			return
		}
		_ = m.ApplyAoEStrike(caster.ID, cmd.To, combat.DefaultAoERadius)
	}
}
