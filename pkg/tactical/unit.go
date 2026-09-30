package tactical

import (
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
)

// UnitIDPlayer0 and UnitIDPlayer1 are the default duel units.
const (
	UnitIDPlayer0 = 101
	UnitIDPlayer1 = 201
)

// Unit is authoritative tactical state for one piece.
type Unit struct {
	ID    uint32
	Owner uint8
	Stats combat.UnitStats
	Pos   board.Coord
}

func defaultUnit(id uint32, owner uint8, typ combat.UnitType, pos board.Coord) Unit {
	hp := fixed.FromInt(100)
	return Unit{
		ID:    id,
		Owner: owner,
		Pos:   pos,
		Stats: combat.UnitStats{
			ID:      id,
			Type:    typ,
			BaseATK: fixed.FromInt(30),
			BaseDEF: fixed.FromInt(10),
			HP:      hp,
			Move:    4,
		},
	}
}
