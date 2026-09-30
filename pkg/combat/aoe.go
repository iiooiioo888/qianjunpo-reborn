package combat

import (
	"sort"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
)

// DefaultAoERadius is the stub splash radius in Chebyshev distance (center cell included).
const DefaultAoERadius = 1

// CollectAoECells returns in-bounds coords with Chebyshev(center, c) <= radius,
// sorted by increasing Y then X for deterministic ordering. radius < 0 yields nil.
func CollectAoECells(center board.Coord, radius int) []board.Coord {
	if radius < 0 {
		return nil
	}
	out := make([]board.Coord, 0, (2*radius+1)*(2*radius+1))
	for y := center.Y - radius; y <= center.Y+radius; y++ {
		for x := center.X - radius; x <= center.X+radius; x++ {
			c := board.Coord{X: x, Y: y}
			if !board.InBounds(c) {
				continue
			}
			if board.Chebyshev(center, c) <= radius {
				out = append(out, c)
			}
		}
	}
	return out
}

// CollectAoETargets returns unit IDs on b within AoECells(center, radius), ascending by ID.
// excludeID is omitted from results (0 = no exclusion). Target collection only; damage is applied in tactical/combat strike helpers.
func CollectAoETargets(b *board.Board, center board.Coord, radius int, excludeID uint32) []uint32 {
	if b == nil {
		return nil
	}
	cells := CollectAoECells(center, radius)
	ids := make([]uint32, 0, len(cells))
	for _, c := range cells {
		id := b.GetUnit(c)
		if id == 0 || id == excludeID {
			continue
		}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
