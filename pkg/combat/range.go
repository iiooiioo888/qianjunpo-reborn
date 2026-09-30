package combat

import "github.com/iiooiioo888/qianjunpo-reborn/pkg/board"

// InAttackRange reports whether target lies within Chebyshev distance [1, maxRange].
func InAttackRange(attacker, target board.Coord, maxRange int) bool {
	d := board.Chebyshev(attacker, target)
	return d >= 1 && d <= maxRange
}

// lineSegment returns the 8-connected straight trace from a to b (inclusive), matching pathfind.ExpandSegment.
func lineSegment(a, b board.Coord) []board.Coord {
	if a.X == b.X && a.Y == b.Y {
		return []board.Coord{a}
	}
	out := []board.Coord{a}
	cur := a
	for cur.X != b.X || cur.Y != b.Y {
		cur = board.Coord{
			X: cur.X + signInt(b.X-cur.X),
			Y: cur.Y + signInt(b.Y-cur.Y),
		}
		out = append(out, cur)
	}
	return out
}

func signInt(v int) int {
	if v > 0 {
		return 1
	}
	if v < 0 {
		return -1
	}
	return 0
}

// AttackLineClear reports whether every cell on the attack trace is terrain-passable and
// no third-party unit occupies a cell on the line (attacker and target IDs are ignored).
func AttackLineClear(b *board.Board, from, to board.Coord, attackerID, targetID uint32) bool {
	seg := lineSegment(from, to)
	if len(seg) == 0 || seg[len(seg)-1] != to {
		return false
	}
	for _, c := range seg {
		if !b.Get(c).Passable {
			return false
		}
		uid := b.GetUnit(c)
		if uid == 0 || uid == attackerID || uid == targetID {
			continue
		}
		return false
	}
	return true
}

// RangedAttackNeedsLine reports whether attack validation must check AttackLineClear.
// Melee (maxRange <= 1) only uses InAttackRange; ranged uses both.
func RangedAttackNeedsLine(maxRange int) bool {
	return maxRange > 1
}
