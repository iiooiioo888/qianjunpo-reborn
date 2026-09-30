package combat

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
)

func TestCollectAoECells_radius1_neighbors(t *testing.T) {
	center := board.Coord{X: 5, Y: 5}
	cells := CollectAoECells(center, DefaultAoERadius)
	if len(cells) != 9 {
		t.Fatalf("expected 9 cells at interior, got %d", len(cells))
	}
	want := map[board.Coord]bool{center: true}
	for _, n := range board.Neighbors(center) {
		want[n] = true
	}
	for _, c := range cells {
		if !want[c] {
			t.Fatalf("unexpected cell %v", c)
		}
	}
	if cells[0].Y > cells[len(cells)-1].Y || (cells[0].Y == cells[len(cells)-1].Y && cells[0].X > cells[len(cells)-1].X) {
		t.Fatal("cells should be sorted by Y then X")
	}
}

func TestCollectAoECells_corner_clipped(t *testing.T) {
	cells := CollectAoECells(board.Coord{X: 0, Y: 0}, 1)
	if len(cells) != 4 {
		t.Fatalf("corner radius 1 want 4 cells, got %d", len(cells))
	}
}

func TestCollectAoECells_negativeRadius(t *testing.T) {
	if CollectAoECells(board.Coord{X: 3, Y: 3}, -1) != nil {
		t.Fatal("negative radius should return nil")
	}
}

func TestCollectAoETargets_neighborsAndExclude(t *testing.T) {
	b := board.New()
	center := board.Coord{X: 10, Y: 10}
	b.SetUnit(center, 100)
	b.SetUnit(board.Coord{X: 11, Y: 10}, 201)
	b.SetUnit(board.Coord{X: 9, Y: 10}, 202)
	b.SetUnit(board.Coord{X: 12, Y: 10}, 203) // Chebyshev 2, out of radius 1

	ids := CollectAoETargets(b, center, DefaultAoERadius, 100)
	if len(ids) != 2 {
		t.Fatalf("want 2 targets in splash, got %v", ids)
	}
	if ids[0] != 201 || ids[1] != 202 {
		t.Fatalf("unexpected ids %v", ids)
	}
}
