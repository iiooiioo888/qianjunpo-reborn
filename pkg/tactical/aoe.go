package tactical

import (
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
)

// CollectAoETargets is a thin match wrapper over combat.CollectAoETargets (AoE stub; no combat apply).
func (m *Match) CollectAoETargets(center board.Coord, radius int, excludeUnitID uint32) []uint32 {
	if m == nil || m.Board == nil {
		return nil
	}
	return combat.CollectAoETargets(m.Board, center, radius, excludeUnitID)
}
