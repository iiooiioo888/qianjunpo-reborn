package pathfind

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
)

func TestUnionFindSameComponent(t *testing.T) {
	uf := newUnionFind(4)
	uf.union(0, 1)
	uf.union(1, 2)
	if !uf.same(0, 2) {
		t.Fatal("expected same")
	}
	if uf.same(0, 3) {
		t.Fatal("expected different")
	}
}

func TestBuildConnectivityUnitBlock(t *testing.T) {
	b := board.New()
	// Single occupied cell does not disconnect the open map; verify it is not in the same
	// component as a cell that requires stepping onto the blocker.
	blocker := board.Coord{9, 9}
	b.SetUnit(blocker, 99)
	conn := BuildConnectivity(b, 0)
	if conn.Reachable(blocker, board.Coord{9, 8}) {
		t.Fatal("should not treat blocked cell as passable")
	}
	connIgnore := BuildConnectivity(b, 99)
	if !connIgnore.Reachable(blocker, board.Coord{9, 8}) {
		t.Fatal("ignored unit tile should stay in component")
	}
}
