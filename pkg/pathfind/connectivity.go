package pathfind

import (
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
)

// Connectivity labels passable cells with union-find for O(1) reachability checks.
type Connectivity struct {
	uf    *unionFind
	index func(board.Coord) int
}

// BuildConnectivity recomputes components from passable cells (units block unless ignored).
func BuildConnectivity(b *board.Board, ignoreUnit uint32) *Connectivity {
	n := board.Size * board.Size
	uf := newUnionFind(n)
	idx := func(c board.Coord) int {
		return c.Y*board.Size + c.X
	}
	for y := 0; y < board.Size; y++ {
		for x := 0; x < board.Size; x++ {
			c := board.Coord{X: x, Y: y}
			if !b.IsPassable(c, ignoreUnit) {
				continue
			}
			for _, nb := range board.Neighbors(c) {
				if b.IsPassable(nb, ignoreUnit) {
					uf.union(idx(c), idx(nb))
				}
			}
		}
	}
	return &Connectivity{uf: uf, index: idx}
}

// Reachable reports whether start and goal share a passable component.
func (c *Connectivity) Reachable(start, goal board.Coord) bool {
	if !board.InBounds(start) || !board.InBounds(goal) {
		return false
	}
	return c.uf.same(c.index(start), c.index(goal))
}
