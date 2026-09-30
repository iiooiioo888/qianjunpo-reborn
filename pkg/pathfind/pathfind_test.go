package pathfind

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
)

func openBoard() *board.Board {
	return board.New()
}

func assertAdjacentPath(t *testing.T, path []board.Coord) {
	t.Helper()
	for i := 1; i < len(path); i++ {
		if board.Chebyshev(path[i-1], path[i]) != 1 {
			t.Fatalf("non-adjacent step %d: %v -> %v", i, path[i-1], path[i])
		}
	}
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
	assertAdjacentPath(t, res.Path)
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

func TestConnectivitySameComponentOpenMap(t *testing.T) {
	b := openBoard()
	conn := BuildConnectivity(b, 0)
	a := board.Coord{0, 0}
	z := board.Coord{18, 18}
	if QuickUnreachable(conn, a, z) {
		t.Fatal("open map should be reachable")
	}
	if !conn.Reachable(a, z) {
		t.Fatal("expected reachable")
	}
}

func TestFindPathFastReject(t *testing.T) {
	b := openBoard()
	for x := 0; x < board.Size; x++ {
		b.SetTerrain(board.Coord{x, 9}, board.TerrainMountain)
	}
	walk := FromBoard(b, 0)
	conn := BuildConnectivity(b, 0)
	res := FindPath(board.Coord{0, 0}, board.Coord{18, 18}, walk, conn, FindOptions{Algo: AlgoAStar})
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
	assertAdjacentPath(t, j.Path)
}

func TestJPSExpandedMatchesAStar(t *testing.T) {
	b := openBoard()
	walk := FromBoard(b, 0)
	cases := []struct {
		from, to board.Coord
	}{
		{board.Coord{0, 0}, board.Coord{18, 18}},
		{board.Coord{3, 1}, board.Coord{15, 17}},
		{board.Coord{0, 18}, board.Coord{18, 0}},
		{board.Coord{9, 9}, board.Coord{12, 4}},
	}
	for _, tc := range cases {
		a := AStar(tc.from, tc.to, walk)
		j := FindPath(tc.from, tc.to, walk, nil, FindOptions{Algo: AlgoJPS})
		if !a.OK || !j.OK {
			t.Fatalf("path failed %v->%v", tc.from, tc.to)
		}
		if PathLength(a.Path) != PathLength(j.Path) {
			t.Fatalf("%v->%v astar=%d jps=%d", tc.from, tc.to, PathLength(a.Path), PathLength(j.Path))
		}
		assertAdjacentPath(t, j.Path)
	}
}

func TestSmoothPathOpenMap(t *testing.T) {
	b := openBoard()
	walk := FromBoard(b, 0)
	start := board.Coord{0, 0}
	goal := board.Coord{6, 6}
	raw := AStar(start, goal, walk)
	if !raw.OK {
		t.Fatal("astar failed")
	}
	sm := SmoothPath(raw.Path, walk)
	if PathLength(sm) != PathLength(raw.Path) {
		t.Fatalf("smooth changed length raw=%d sm=%d", PathLength(raw.Path), PathLength(sm))
	}
	assertAdjacentPath(t, sm)
	if len(sm) > len(raw.Path) {
		t.Fatalf("expected fewer or equal waypoints got raw=%d sm=%d", len(raw.Path), len(sm))
	}
}

func TestHasLineOfSightBlocked(t *testing.T) {
	b := openBoard()
	b.SetTerrain(board.Coord{3, 3}, board.TerrainMountain)
	walk := FromBoard(b, 0)
	if HasLineOfSight(board.Coord{0, 0}, board.Coord{6, 6}, walk) {
		t.Fatal("expected blocked LOS")
	}
	if !HasLineOfSight(board.Coord{0, 0}, board.Coord{6, 0}, walk) {
		t.Fatal("expected clear horizontal LOS")
	}
}

func TestExpandSegment(t *testing.T) {
	seg := ExpandSegment(board.Coord{0, 0}, board.Coord{3, 2})
	if len(seg) != 4 {
		t.Fatalf("len=%d", len(seg))
	}
	if seg[len(seg)-1].X != 3 || seg[len(seg)-1].Y != 2 {
		t.Fatal("bad end", seg[len(seg)-1])
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

func BenchmarkAStarOpen200(b *testing.B) {
	brd := openBoard()
	walk := FromBoard(brd, 0)
	start := board.Coord{0, 0}
	goal := board.Coord{14, 14} // ~14 cheb steps, ~200 cells region
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AStar(start, goal, walk)
	}
}

func BenchmarkJPSOpen200(b *testing.B) {
	brd := openBoard()
	walk := FromBoard(brd, 0)
	start := board.Coord{0, 0}
	goal := board.Coord{14, 14}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		JPS(start, goal, walk)
	}
}

func BenchmarkFindPathSmoothOpen(b *testing.B) {
	brd := openBoard()
	walk := FromBoard(brd, 0)
	start := board.Coord{0, 0}
	goal := board.Coord{14, 14}
	opt := FindOptions{Algo: AlgoAStar, Smooth: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindPath(start, goal, walk, nil, opt)
	}
}
