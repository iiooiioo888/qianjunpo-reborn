package tactical

import (
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
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

func defaultUnit(cfg combat.Config, id uint32, owner uint8, typ combat.UnitType, pos board.Coord) Unit {
	stats, err := cfg.CatalogStats(id, typ)
	if err != nil {
		panic(err)
	}
	return Unit{
		ID:    id,
		Owner: owner,
		Pos:   pos,
		Stats: stats,
	}
}
