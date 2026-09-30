package combat

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
)

func TestInAttackRange(t *testing.T) {
	a := board.Coord{5, 5}
	if !InAttackRange(a, board.Coord{6, 5}, 1) {
		t.Fatal("adjacent should be in melee range")
	}
	if InAttackRange(a, board.Coord{7, 5}, 1) {
		t.Fatal("distance 2 out of melee range")
	}
	if !InAttackRange(a, board.Coord{7, 7}, 2) {
		t.Fatal("chebyshev 2 should be in range 2")
	}
	if InAttackRange(a, a, 3) {
		t.Fatal("zero distance invalid")
	}
}

func TestAttackLineClear(t *testing.T) {
	b := board.New()
	from := board.Coord{0, 0}
	to := board.Coord{3, 0}
	if !AttackLineClear(b, from, to, 1, 2) {
		t.Fatal("open line expected clear")
	}
	blocker := board.Coord{1, 0}
	b.SetUnit(blocker, 99)
	if AttackLineClear(b, from, to, 1, 2) {
		t.Fatal("unit on line should block")
	}
	if !AttackLineClear(b, from, to, 1, 99) {
		t.Fatal("ignored unit on line should not block")
	}
	b.ClearUnit(blocker)
	b.SetPassable(board.Coord{2, 0}, false)
	if AttackLineClear(b, from, to, 1, 2) {
		t.Fatal("impassable terrain should block")
	}
}

func TestRangedAttackNeedsLine(t *testing.T) {
	if RangedAttackNeedsLine(1) {
		t.Fatal("melee should not need line check flag")
	}
	if !RangedAttackNeedsLine(3) {
		t.Fatal("archer range should need line")
	}
}
