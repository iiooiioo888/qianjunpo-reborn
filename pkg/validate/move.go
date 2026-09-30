package validate

import (
	"fmt"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/pathfind"
)

// MoveError describes why a move was rejected.
type MoveError struct {
	Code    string
	Message string
}

func (e MoveError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

const (
	CodeOutOfBounds = "OUT_OF_BOUNDS"
	CodeWrongStart  = "WRONG_START"
	CodeExceedsMove = "EXCEEDS_MOVE"
	CodeNotAdjacent = "NOT_ADJACENT"
	CodeBlocked     = "BLOCKED"
	CodeOccupied    = "OCCUPIED"
	CodeNoPath      = "NO_PATH"
	CodePathTooLong = "PATH_TOO_LONG"
)

// MoveRequest is server-authoritative move validation input.
type MoveRequest struct {
	UnitID     uint32
	From       board.Coord
	To         board.Coord
	MovePoints int           // max Chebyshev steps along path
	Path       []board.Coord // optional explicit path; if empty, pathfinder fills it
}

// Validator checks moves against the board.
type Validator struct {
	Board *board.Board
}

// NewValidator creates a validator for b.
func NewValidator(b *board.Board) *Validator {
	return &Validator{Board: b}
}

// ValidateMove enforces start, range, adjacency along path, passability, and no unit blocking.
func (v *Validator) ValidateMove(req MoveRequest) ([]board.Coord, error) {
	b := v.Board
	if !board.InBounds(req.From) || !board.InBounds(req.To) {
		return nil, MoveError{Code: CodeOutOfBounds, Message: "from or to off board"}
	}
	if b.GetUnit(req.From) != req.UnitID {
		return nil, MoveError{Code: CodeWrongStart, Message: "unit not at from"}
	}
	if req.UnitID == 0 {
		return nil, MoveError{Code: CodeWrongStart, Message: "invalid unit id"}
	}
	if b.GetUnit(req.To) != 0 && b.GetUnit(req.To) != req.UnitID {
		return nil, MoveError{Code: CodeOccupied, Message: "destination occupied"}
	}

	path := req.Path
	if len(path) == 0 {
		walk := pathfind.FromBoard(b, req.UnitID)
		conn := pathfind.BuildConnectivity(b, req.UnitID)
		if pathfind.QuickUnreachable(conn, req.From, req.To) {
			return nil, MoveError{Code: CodeNoPath, Message: "no route"}
		}
		res := pathfind.FindPath(req.From, req.To, walk, conn, pathfind.FindOptions{
			Algo:   pathfind.AlgoAStar,
			Smooth: true,
		})
		if !res.OK {
			return nil, MoveError{Code: CodeNoPath, Message: "pathfinding failed"}
		}
		path = res.Path
	}

	if len(path) < 1 {
		return nil, MoveError{Code: CodeNoPath, Message: "empty path"}
	}
	if path[0].X != req.From.X || path[0].Y != req.From.Y {
		return nil, MoveError{Code: CodeWrongStart, Message: "path does not start at from"}
	}
	end := path[len(path)-1]
	if end.X != req.To.X || end.Y != req.To.Y {
		return nil, MoveError{Code: CodeWrongStart, Message: "path does not end at to"}
	}

	steps := pathfind.PathLength(path)
	if steps > req.MovePoints {
		return nil, MoveError{Code: CodeExceedsMove, Message: fmt.Sprintf("need %d steps, have %d", steps, req.MovePoints)}
	}

	for i := 1; i < len(path); i++ {
		prev, cur := path[i-1], path[i]
		if board.Chebyshev(prev, cur) != 1 {
			return nil, MoveError{Code: CodeNotAdjacent, Message: "non-adjacent step"}
		}
		if !b.IsPassable(cur, req.UnitID) {
			return nil, MoveError{Code: CodeBlocked, Message: "cell blocked"}
		}
	}

	if steps > req.MovePoints {
		return nil, MoveError{Code: CodePathTooLong, Message: "path too long"}
	}
	return path, nil
}

// ApplyMove updates the board after successful validation.
func (v *Validator) ApplyMove(req MoveRequest, path []board.Coord) error {
	req.Path = path
	if _, err := v.ValidateMove(req); err != nil {
		return err
	}
	v.Board.ClearUnit(req.From)
	if !v.Board.SetUnit(req.To, req.UnitID) {
		return MoveError{Code: CodeOccupied, Message: "cannot place unit"}
	}
	return nil
}
