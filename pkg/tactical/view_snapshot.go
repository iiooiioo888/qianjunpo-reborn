package tactical

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
)

const viewSnapshotSchemaVersion = 1

// ViewCell is display-layer terrain occupancy (mirrors pkg/board.Cell, not authoritative).
type ViewCell struct {
	Terrain  uint8 `json:"terrain"`
	Passable bool  `json:"passable"`
	UnitID   uint32 `json:"unitId"`
}

// ViewUnit is a tactical piece for client placeholders.
type ViewUnit struct {
	ID    uint32 `json:"id"`
	Owner uint8  `json:"owner"`
	Type  uint8  `json:"type"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	HP    int64  `json:"hp"`
	MaxHP int64  `json:"maxHp"`
}

// ViewSkillCast is display-layer metadata for the latest KindSkill resolution (not in StateHash).
type ViewSkillCast struct {
	SkillID      uint16 `json:"skillId"`
	CasterUnitID uint32 `json:"casterUnitId"`
	TargetX      int    `json:"targetX"`
	TargetY      int    `json:"targetY"`
	LockstepFrame uint64 `json:"lockstepFrame"`
}

// ViewSnapshot is JSON consumed by the Cocos display layer (logic/display separation).
type ViewSnapshot struct {
	SchemaVersion     int        `json:"schemaVersion"`
	BoardSize         int        `json:"boardSize"`
	Seed              uint64     `json:"seed"`
	LockstepFrame     uint64     `json:"lockstepFrame"`
	TimeFlowRateParts uint32     `json:"timeFlowRateParts"`
	InitialStateHash  string     `json:"initialStateHash"`
	Cells             [][]ViewCell `json:"cells"`
	Units             []ViewUnit `json:"units"`
	LastSkillCast     *ViewSkillCast `json:"lastSkillCast,omitempty"`
}

// MatchToViewSnapshot exports current match state for local client preview.
func MatchToViewSnapshot(m *Match, timeFlowRateParts uint32) ViewSnapshot {
	if timeFlowRateParts == 0 {
		timeFlowRateParts = 10000
	}
	cells := make([][]ViewCell, board.Size)
	for y := 0; y < board.Size; y++ {
		row := make([]ViewCell, board.Size)
		for x := 0; x < board.Size; x++ {
			c := m.Board.Get(board.Coord{X: x, Y: y})
			row[x] = ViewCell{
				Terrain:  uint8(c.Terrain),
				Passable: c.Passable,
				UnitID:   c.UnitID,
			}
		}
		cells[y] = row
	}
	ids := make([]uint32, 0, len(m.Units))
	for id := range m.Units {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	units := make([]ViewUnit, 0, len(ids))
	for _, id := range ids {
		u := m.Units[id]
		if u == nil {
			continue
		}
		units = append(units, ViewUnit{
			ID: u.ID, Owner: u.Owner, Type: uint8(u.Stats.Type),
			X: u.Pos.X, Y: u.Pos.Y,
			HP: u.Stats.HP.Raw(), MaxHP: 100,
		})
	}
	var lastCast *ViewSkillCast
	if rec := m.LastSkillCast(); rec != nil {
		lastCast = &ViewSkillCast{
			SkillID:       uint16(rec.SkillID),
			CasterUnitID:  rec.CasterUnitID,
			TargetX:       rec.Target.X,
			TargetY:       rec.Target.Y,
			LockstepFrame: rec.Frame,
		}
	}
	return ViewSnapshot{
		SchemaVersion:     viewSnapshotSchemaVersion,
		BoardSize:         board.Size,
		Seed:              m.Seed,
		LockstepFrame:     m.Frame,
		TimeFlowRateParts: timeFlowRateParts,
		InitialStateHash:  fmt.Sprintf("%016x", m.initial),
		Cells:             cells,
		Units:             units,
		LastSkillCast:     lastCast,
	}
}

// InitialViewSnapshot returns the default duel layout (seed 0xcafe matches make match-play).
func InitialViewSnapshot(seed uint64) ViewSnapshot {
	return MatchToViewSnapshot(NewMatch(seed), 10000)
}

// MarshalViewSnapshotJSON pretty-prints a view snapshot for checked-in mock data.
func MarshalViewSnapshotJSON(v ViewSnapshot) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// UnitTypeName maps combat unit type to a display token (client may localize).
func UnitTypeName(t combat.UnitType) string {
	switch t {
	case combat.UnitInfantry:
		return "infantry"
	case combat.UnitArcher:
		return "archer"
	case combat.UnitCavalry:
		return "cavalry"
	default:
		return "unknown"
	}
}
