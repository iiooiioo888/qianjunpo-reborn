package pathfind

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
)

func openBoard() *board.Board {
	return board.New()
}

func TestAStarStraightPath(t *testing.T) {
	b := openBoard()
	walk := FromBoard(b, 0)
	start := board.Coord{0, 0}
	goal := board.Coord{5, 0}
	res := AStar(start, goal, walk)
	if !res.OK || len(res.Path) != 6 {
		t.Fatalf("path: ok=%v len=%d", res.OK, len(res.Path))
	}
}

func TestAStarUnreachable(t *testing.T) {
	b := openBoard()
	for x := 0; x < board.Size; x++ {
		b.SetTerrain(board.Coord{x, 9}, board.TerrainMountain)
	}
	walk := FromBoard(b, 0)
	start := board.Coord{0, 0}
	goal := board.Coord{18, 18}
	conn := BuildConnectivity(b, 0)
	if !QuickUnreachable(conn, start, goal) {
		t.Fatal("connectivity should detect unreachable")
	}
	res := AStar(start, goal, walk)
	if res.OK {
		t.Fatal("expected no path")
	}
}

func TestJPSMatchesAStar(t *testing.T) {
	b := openBoard()
	walk := FromBoard(b, 0)
	start := board.Coord{2, 2}
	goal := board.Coord{10, 12}
	a := AStar(start, goal, walk)
	j := JPS(start, goal, walk)
	if !a.OK || !j.OK {
		t.Fatal("both should find path")
	}
	if PathLength(a.Path) != PathLength(j.Path) {
		t.Fatalf("length mismatch astar=%d jps=%d", PathLength(a.Path), PathLength(j.Path))
	}
}

func TestPerfSmoke(t *testing.T) {
	b := openBoard()
	walk := FromBoard(b, 0)
	start := board.Coord{0, 0}
	goal := board.Coord{18, 18}
	res, expanded := AStarWithBudget(start, goal, walk, SmokeBudget)
	if !res.OK {
		t.Fatalf("smoke path failed expanded=%d", expanded)
	}
	if expanded > SmokeBudget {
		t.Fatalf("exceeded budget")
	}
}
