package pathfind

import "github.com/iiooiioo888/qianjunpo-reborn/pkg/board"

// HasLineOfSight reports whether every cell on the Chebyshev straight trace from a to b is walkable.
// The trace matches ExpandSegment: each step moves one cell diagonally or orthogonally toward b.
func HasLineOfSight(a, b board.Coord, walk Walkable) bool {
	if !walk(a) || !walk(b) {
		return false
	}
	seg := ExpandSegment(a, b)
	for _, c := range seg {
		if !walk(c) {
			return false
		}
	}
	return true
}

// ExpandSegment returns an adjacent-step path from a to b along the Chebyshev shortest direction.
// If a and b are not on a straight 8-connected line, the result may not end at b.
func ExpandSegment(a, b board.Coord) []board.Coord {
	if a.X == b.X && a.Y == b.Y {
		return []board.Coord{a}
	}
	out := []board.Coord{a}
	cur := a
	for cur.X != b.X || cur.Y != b.Y {
		cur = board.Coord{
			X: cur.X + sign(b.X-cur.X),
			Y: cur.Y + sign(b.Y-cur.Y),
		}
		out = append(out, cur)
	}
	return out
}

// ExpandJumpPath turns a jump-point polyline into adjacent steps (required for move validation).
func ExpandJumpPath(jumpPath []board.Coord) []board.Coord {
	if len(jumpPath) == 0 {
		return nil
	}
	out := []board.Coord{jumpPath[0]}
	for i := 1; i < len(jumpPath); i++ {
		seg := ExpandSegment(out[len(out)-1], jumpPath[i])
		out = append(out, seg[1:]...)
	}
	return out
}

// SmoothPath merges collinear visibility using line-of-sight (Bresenham-style Chebyshev trace).
// Output remains adjacent-step; PathLength is never greater than the input.
func SmoothPath(path []board.Coord, walk Walkable) []board.Coord {
	if len(path) <= 2 {
		return append([]board.Coord(nil), path...)
	}
	out := []board.Coord{path[0]}
	i := 0
	for i < len(path)-1 {
		j := len(path) - 1
		for j > i+1 && !HasLineOfSight(path[i], path[j], walk) {
			j--
		}
		seg := ExpandSegment(path[i], path[j])
		out = append(out, seg[1:]...)
		i = j
	}
	return out
}
