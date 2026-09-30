package board

import "testing"

func TestChebyshevAndNeighbors(t *testing.T) {
	a := Coord{0, 0}
	b := Coord{3, 5}
	if Chebyshev(a, b) != 5 {
		t.Fatalf("chebyshev want 5 got %d", Chebyshev(a, b))
	}
	ns := Neighbors(Coord{1, 1})
	if len(ns) != 8 {
		t.Fatalf("expected 8 neighbors, got %d", len(ns))
	}
	nsCorner := Neighbors(Coord{0, 0})
	if len(nsCorner) != 3 {
		t.Fatalf("corner neighbors want 3 got %d", len(nsCorner))
	}
}

func TestUnitPlacement(t *testing.T) {
	b := New()
	c := Coord{5, 5}
	if !b.SetUnit(c, 42) {
		t.Fatal("set unit")
	}
	if b.GetUnit(c) != 42 {
		t.Fatal("get unit")
	}
	if b.IsPassable(c, 0) {
		t.Fatal("blocked by unit")
	}
	if !b.IsPassable(c, 42) {
		t.Fatal("self passable")
	}
	b.ClearUnit(c)
	if b.GetUnit(c) != 0 {
		t.Fatal("cleared")
	}
}

func TestTerrainPassability(t *testing.T) {
	b := New()
	m := Coord{2, 2}
	b.SetTerrain(m, TerrainMountain)
	if b.IsPassable(m, 0) {
		t.Fatal("mountain impassable")
	}
	b.SetPassable(m, true)
	if !b.IsPassable(m, 0) {
		t.Fatal("override passable")
	}
}
