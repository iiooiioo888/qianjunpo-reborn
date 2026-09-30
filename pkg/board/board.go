// Package board provides the authoritative 19×19 war-chess grid.
package board

const Size = 19

// Terrain classifies a cell's map feature.
type Terrain uint8

const (
	TerrainPlain    Terrain = 0
	TerrainMountain Terrain = 1
	TerrainForest   Terrain = 2
	TerrainRiver    Terrain = 3
	TerrainCity     Terrain = 4
	TerrainPass     Terrain = 5
)

// Effect flags on a cell (stackable bitmask).
type Effect uint16

const (
	EffectNone Effect = 0
)

// Cell is one board tile.
type Cell struct {
	Terrain  Terrain
	Passable bool
	UnitID   uint32 // 0 = empty
	Effects  Effect
}

// Coord is an integer grid position.
type Coord struct {
	X int
	Y int
}

// InBounds reports whether c lies on the board.
func InBounds(c Coord) bool {
	return c.X >= 0 && c.X < Size && c.Y >= 0 && c.Y < Size
}

// Chebyshev returns max(|dx|,|dy|) between two coords.
func Chebyshev(a, b Coord) int {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx
	}
	return dy
}

// Neighbors returns up to eight adjacent in-bounds coords (N, NE, E, SE, S, SW, W, NW).
func Neighbors(c Coord) []Coord {
	dirs := [8]Coord{
		{0, -1}, {1, -1}, {1, 0}, {1, 1},
		{0, 1}, {-1, 1}, {-1, 0}, {-1, -1},
	}
	out := make([]Coord, 0, 8)
	for _, d := range dirs {
		nc := Coord{c.X + d.X, c.Y + d.Y}
		if InBounds(nc) {
			out = append(out, nc)
		}
	}
	return out
}

// Board is a fixed-size 19×19 grid.
type Board struct {
	cells [Size][Size]Cell
}

// New creates a board with default passability per terrain.
func New() *Board {
	b := &Board{}
	for y := 0; y < Size; y++ {
		for x := 0; x < Size; x++ {
			b.cells[y][x] = defaultCell(TerrainPlain)
		}
	}
	return b
}

func defaultCell(t Terrain) Cell {
	pass := true
	switch t {
	case TerrainMountain, TerrainRiver:
		pass = false
	case TerrainPlain, TerrainForest, TerrainCity, TerrainPass:
		pass = true
	}
	return Cell{Terrain: t, Passable: pass, UnitID: 0, Effects: EffectNone}
}

// Get returns the cell at c (caller must ensure InBounds).
func (b *Board) Get(c Coord) Cell {
	return b.cells[c.Y][c.X]
}

// SetTerrain sets terrain and default passability at c.
func (b *Board) SetTerrain(c Coord, t Terrain) {
	if !InBounds(c) {
		return
	}
	b.cells[c.Y][c.X] = defaultCell(t)
}

// SetPassable overrides passability (e.g. bridge over river).
func (b *Board) SetPassable(c Coord, passable bool) {
	if !InBounds(c) {
		return
	}
	b.cells[c.Y][c.X].Passable = passable
}

// IsPassable reports whether a unit may enter c (terrain + optional unit blocking).
func (b *Board) IsPassable(c Coord, ignoreUnitID uint32) bool {
	if !InBounds(c) {
		return false
	}
	cell := b.cells[c.Y][c.X]
	if !cell.Passable {
		return false
	}
	if cell.UnitID != 0 && cell.UnitID != ignoreUnitID {
		return false
	}
	return true
}

// GetUnit returns the unit id at c, or 0 if empty.
func (b *Board) GetUnit(c Coord) uint32 {
	if !InBounds(c) {
		return 0
	}
	return b.cells[c.Y][c.X].UnitID
}

// SetUnit places unitID at c (must be empty or same id).
func (b *Board) SetUnit(c Coord, unitID uint32) bool {
	if !InBounds(c) || unitID == 0 {
		return false
	}
	if b.cells[c.Y][c.X].UnitID != 0 && b.cells[c.Y][c.X].UnitID != unitID {
		return false
	}
	b.cells[c.Y][c.X].UnitID = unitID
	return true
}

// ClearUnit removes any unit at c.
func (b *Board) ClearUnit(c Coord) {
	if !InBounds(c) {
		return
	}
	b.cells[c.Y][c.X].UnitID = 0
}

// Hash writes deterministic cell state into the hasher callback pattern.
func (b *Board) Hash(h func(uint64)) {
	for y := 0; y < Size; y++ {
		for x := 0; x < Size; x++ {
			c := b.cells[y][x]
			h(uint64(c.Terrain))
			if c.Passable {
				h(1)
			} else {
				h(0)
			}
			h(uint64(c.UnitID))
			h(uint64(c.Effects))
		}
	}
}
