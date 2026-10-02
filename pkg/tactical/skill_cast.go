package tactical

import (
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
)

// SkillCastRecord is authoritative metadata for the most recent resolved KindSkill (display + replay hooks).
type SkillCastRecord struct {
	SkillID      combat.SkillID
	CasterUnitID uint32
	Target       board.Coord
	Frame        uint64
}

// LastSkillCast returns the latest skill resolved on the lockstep path (nil if none yet).
func (m *Match) LastSkillCast() *SkillCastRecord {
	if m == nil || m.lastSkillCast == nil {
		return nil
	}
	cp := *m.lastSkillCast
	return &cp
}

func (m *Match) recordSkillCast(cmd Command, caster *Unit) {
	if m == nil || caster == nil || cmd.SkillID == 0 {
		return
	}
	m.lastSkillCast = &SkillCastRecord{
		SkillID:      cmd.SkillID,
		CasterUnitID: caster.ID,
		Target:       cmd.To,
		Frame:        m.Frame,
	}
}
