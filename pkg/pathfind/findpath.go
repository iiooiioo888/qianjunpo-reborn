package pathfind

import "github.com/iiooiioo888/qianjunpo-reborn/pkg/board"

// Algorithm selects the search backend for FindPath.
type Algorithm int

const (
	// AlgoAStar is canonical optimal search on the 8-connected grid.
	AlgoAStar Algorithm = iota
	// AlgoJPS uses jump-point search then ExpandJumpPath for adjacent steps.
	AlgoJPS
)

// FindOptions configures FindPath.
type FindOptions struct {
	Algo   Algorithm
	Smooth bool
}

// FindPath optionally prechecks connectivity, runs A* or JPS, and optionally smooths.
// conn may be nil to skip the fast unreachable check.
func FindPath(start, goal board.Coord, walk Walkable, conn *Connectivity, opt FindOptions) PathResult {
	if conn != nil && QuickUnreachable(conn, start, goal) {
		return PathResult{OK: false}
	}
	var res PathResult
	switch opt.Algo {
	case AlgoJPS:
		res = JPS(start, goal, walk)
	default:
		res = AStar(start, goal, walk)
	}
	if !res.OK {
		return res
	}
	if opt.Smooth {
		res.Path = SmoothPath(res.Path, walk)
	}
	return res
}
