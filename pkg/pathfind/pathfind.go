package pathfind

import (
	"container/heap"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
)

// Walkable answers whether a coord can be entered.
type Walkable func(c board.Coord) bool

// FromBoard builds walkability using board passability and optional unit ignore.
func FromBoard(b *board.Board, ignoreUnit uint32) Walkable {
	return func(c board.Coord) bool {
		return b.IsPassable(c, ignoreUnit)
	}
}

// PathResult holds a found path or failure.
type PathResult struct {
	Path []board.Coord
	OK   bool
}

type pqItem struct {
	c    board.Coord
	f    int
	g    int
	from board.Coord
}

type priorityQueue []*pqItem

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].f < pq[j].f }
func (pq priorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }
func (pq *priorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*pqItem))
}
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}

// AStar finds a shortest path under Chebyshev step costs (each step cost 1).
func AStar(start, goal board.Coord, walk Walkable) PathResult {
	if !walk(start) || !walk(goal) {
		return PathResult{OK: false}
	}
	if start.X == goal.X && start.Y == goal.Y {
		return PathResult{Path: []board.Coord{start}, OK: true}
	}

	open := &priorityQueue{}
	heap.Init(open)
	heap.Push(open, &pqItem{c: start, g: 0, f: cheb(start, goal), from: board.Coord{-1, -1}})

	gScore := map[int]int{}
	gScore[key(start)] = 0
	cameFrom := map[int]board.Coord{}
	closed := map[int]bool{}

	for open.Len() > 0 {
		cur := heap.Pop(open).(*pqItem)
		k := key(cur.c)
		if closed[k] {
			continue
		}
		closed[k] = true
		if cur.c.X == goal.X && cur.c.Y == goal.Y {
			return PathResult{Path: reconstruct(cameFrom, start, goal), OK: true}
		}
		for _, nb := range board.Neighbors(cur.c) {
			if !walk(nb) {
				continue
			}
			ng := cur.g + stepCost(cur.c, nb)
			nk := key(nb)
			if prev, ok := gScore[nk]; ok && ng >= prev {
				continue
			}
			gScore[nk] = ng
			cameFrom[nk] = cur.c
			heap.Push(open, &pqItem{c: nb, g: ng, f: ng + cheb(nb, goal), from: cur.c})
		}
	}
	return PathResult{OK: false}
}

func cheb(a, b board.Coord) int {
	return board.Chebyshev(a, b)
}

func stepCost(a, b board.Coord) int {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	if dx == 0 || dy == 0 || dx == dy {
		return 1
	}
	return 1
}

func key(c board.Coord) int {
	return c.Y*board.Size + c.X
}

func reconstruct(cameFrom map[int]board.Coord, start, goal board.Coord) []board.Coord {
	path := []board.Coord{goal}
	cur := goal
	for {
		if cur.X == start.X && cur.Y == start.Y {
			break
		}
		prev := cameFrom[key(cur)]
		path = append(path, prev)
		cur = prev
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// JPS finds a path using jump-point search on an 8-connected grid.
func JPS(start, goal board.Coord, walk Walkable) PathResult {
	if !walk(start) || !walk(goal) {
		return PathResult{OK: false}
	}
	if start.X == goal.X && start.Y == goal.Y {
		return PathResult{Path: []board.Coord{start}, OK: true}
	}

	open := &priorityQueue{}
	heap.Init(open)
	heap.Push(open, &pqItem{c: start, g: 0, f: cheb(start, goal)})

	gScore := map[int]int{key(start): 0}
	cameFrom := map[int]board.Coord{}

	for open.Len() > 0 {
		cur := heap.Pop(open).(*pqItem)
		if cur.c.X == goal.X && cur.c.Y == goal.Y {
			return PathResult{Path: reconstruct(cameFrom, start, goal), OK: true}
		}
		dirs := prunedDirs(cur.c, cameFrom, start, walk)
		for _, d := range dirs {
			jp := jump(cur.c, d, goal, walk)
			if jp.X < 0 {
				continue
			}
			ng := gScore[key(cur.c)] + cheb(cur.c, jp)
			jk := key(jp)
			if prev, ok := gScore[jk]; ok && ng >= prev {
				continue
			}
			gScore[jk] = ng
			cameFrom[jk] = cur.c
			heap.Push(open, &pqItem{c: jp, g: ng, f: ng + cheb(jp, goal)})
		}
	}
	return PathResult{OK: false}
}

var allDirs = [8]board.Coord{
	{0, -1}, {1, 0}, {0, 1}, {-1, 0},
	{1, -1}, {1, 1}, {-1, 1}, {-1, -1},
}

func prunedDirs(c board.Coord, cameFrom map[int]board.Coord, start board.Coord, walk Walkable) []board.Coord {
	if c.X == start.X && c.Y == start.Y {
		out := make([]board.Coord, 0, 8)
		for _, d := range allDirs {
			nb := board.Coord{c.X + d.X, c.Y + d.Y}
			if walk(nb) {
				out = append(out, d)
			}
		}
		return out
	}
	parent := cameFrom[key(c)]
	dx := sign(c.X - parent.X)
	dy := sign(c.Y - parent.Y)
	var dirs []board.Coord
	if dx != 0 && dy != 0 {
		if walk(board.Coord{c.X, c.Y + dy}) {
			dirs = append(dirs, board.Coord{0, dy})
		}
		if walk(board.Coord{c.X + dx, c.Y}) {
			dirs = append(dirs, board.Coord{dx, 0})
		}
		if walk(board.Coord{c.X + dx, c.Y + dy}) {
			dirs = append(dirs, board.Coord{dx, dy})
		}
		forced := jumpForcedDiagonal(c, board.Coord{dx, dy}, walk)
		dirs = append(dirs, forced...)
	} else {
		if dx != 0 {
			if walk(board.Coord{c.X + dx, c.Y}) {
				dirs = append(dirs, board.Coord{dx, 0})
			}
			forced := jumpForcedStraight(c, board.Coord{dx, 0}, walk)
			dirs = append(dirs, forced...)
		} else {
			if walk(board.Coord{c.X, c.Y + dy}) {
				dirs = append(dirs, board.Coord{0, dy})
			}
			forced := jumpForcedStraight(c, board.Coord{0, dy}, walk)
			dirs = append(dirs, forced...)
		}
	}
	return dirs
}

func jumpForcedStraight(c, d board.Coord, walk Walkable) []board.Coord {
	var out []board.Coord
	if d.X != 0 {
		if !walk(board.Coord{c.X, c.Y + 1}) && walk(board.Coord{c.X + d.X, c.Y + 1}) {
			out = append(out, board.Coord{d.X, 1})
		}
		if !walk(board.Coord{c.X, c.Y - 1}) && walk(board.Coord{c.X + d.X, c.Y - 1}) {
			out = append(out, board.Coord{d.X, -1})
		}
	} else {
		if !walk(board.Coord{c.X + 1, c.Y}) && walk(board.Coord{c.X + 1, c.Y + d.Y}) {
			out = append(out, board.Coord{1, d.Y})
		}
		if !walk(board.Coord{c.X - 1, c.Y}) && walk(board.Coord{c.X - 1, c.Y + d.Y}) {
			out = append(out, board.Coord{-1, d.Y})
		}
	}
	return out
}

func jumpForcedDiagonal(c, d board.Coord, walk Walkable) []board.Coord {
	var out []board.Coord
	if !walk(board.Coord{c.X - d.X, c.Y}) && walk(board.Coord{c.X - d.X, c.Y + d.Y}) {
		out = append(out, board.Coord{-d.X, d.Y})
	}
	if !walk(board.Coord{c.X, c.Y - d.Y}) && walk(board.Coord{c.X + d.X, c.Y - d.Y}) {
		out = append(out, board.Coord{d.X, -d.Y})
	}
	return out
}

func jump(c, d, goal board.Coord, walk Walkable) board.Coord {
	nx, ny := c.X, c.Y
	for {
		nx += d.X
		ny += d.Y
		nc := board.Coord{nx, ny}
		if !board.InBounds(nc) || !walk(nc) {
			return board.Coord{-1, -1}
		}
		if nx == goal.X && ny == goal.Y {
			return nc
		}
		if hasForced(nc, d, walk) {
			return nc
		}
		if d.X != 0 && d.Y != 0 {
			if jump(nc, board.Coord{d.X, 0}, goal, walk).X >= 0 || jump(nc, board.Coord{0, d.Y}, goal, walk).X >= 0 {
				return nc
			}
		}
	}
}

func hasForced(c, d board.Coord, walk Walkable) bool {
	if d.X != 0 && d.Y != 0 {
		return (!walk(board.Coord{c.X - d.X, c.Y}) && walk(board.Coord{c.X - d.X, c.Y + d.Y})) ||
			(!walk(board.Coord{c.X, c.Y - d.Y}) && walk(board.Coord{c.X + d.X, c.Y - d.Y}))
	}
	return len(jumpForcedStraight(c, d, walk)) > 0
}

func sign(v int) int {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}

// QuickUnreachable uses connectivity before running full search.
func QuickUnreachable(conn *Connectivity, start, goal board.Coord) bool {
	return !conn.Reachable(start, goal)
}

// PathLength returns Chebyshev path length in steps.
func PathLength(path []board.Coord) int {
	if len(path) <= 1 {
		return 0
	}
	sum := 0
	for i := 1; i < len(path); i++ {
		sum += cheb(path[i-1], path[i])
	}
	return sum
}

// SmokeBudget is a soft node-expansion limit for perf tests.
const SmokeBudget = 5000

// AStarWithBudget stops after expanding budget nodes (for smoke tests).
func AStarWithBudget(start, goal board.Coord, walk Walkable, budget int) (PathResult, int) {
	if !walk(start) || !walk(goal) {
		return PathResult{OK: false}, 0
	}
	open := &priorityQueue{}
	heap.Init(open)
	heap.Push(open, &pqItem{c: start, g: 0, f: cheb(start, goal)})
	gScore := map[int]int{key(start): 0}
	cameFrom := map[int]board.Coord{}
	expanded := 0
	for open.Len() > 0 && expanded < budget {
		cur := heap.Pop(open).(*pqItem)
		expanded++
		if cur.c.X == goal.X && cur.c.Y == goal.Y {
			return PathResult{Path: reconstruct(cameFrom, start, goal), OK: true}, expanded
		}
		for _, nb := range board.Neighbors(cur.c) {
			if !walk(nb) {
				continue
			}
			ng := cur.g + 1
			nk := key(nb)
			if prev, ok := gScore[nk]; ok && ng >= prev {
				continue
			}
			gScore[nk] = ng
			cameFrom[nk] = cur.c
			heap.Push(open, &pqItem{c: nb, g: ng, f: ng + cheb(nb, goal)})
		}
	}
	if expanded >= budget {
		return PathResult{OK: false}, expanded
	}
	return PathResult{OK: false}, expanded
}

// DiagonalBlocked optionally blocks corner-cutting (not used by default).
func DiagonalBlocked(a, b board.Coord, walk Walkable) bool {
	if a.X == b.X || a.Y == b.Y {
		return false
	}
	return !walk(board.Coord{b.X, a.Y}) || !walk(board.Coord{a.X, b.Y})
}
