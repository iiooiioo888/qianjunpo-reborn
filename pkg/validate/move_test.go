package validate

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
)

func TestValidateMoveSuccess(t *testing.T) {
	b := board.New()
	from := board.Coord{1, 1}
	to := board.Coord{3, 2}
	b.SetUnit(from, 7)
	v := NewValidator(b)
	path, err := v.ValidateMove(MoveRequest{
		UnitID: 7, From: from, To: to, MovePoints: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(path) < 2 {
		t.Fatal("expected path")
	}
}

func TestValidateMoveBlocked(t *testing.T) {
	b := board.New()
	from := board.Coord{5, 5}
	to := board.Coord{7, 5}
	b.SetUnit(from, 1)
	// Block straight line and simple detours in a small box.
	for x := 4; x <= 8; x++ {
		for y := 4; y <= 6; y++ {
			if x == 5 && y == 5 {
				continue
			}
			if x == 7 && y == 5 {
				continue
			}
			b.SetUnit(board.Coord{x, y}, 2)
		}
	}
	v := NewValidator(b)
	_, err := v.ValidateMove(MoveRequest{
		UnitID: 1, From: from, To: to, MovePoints: 20,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateInsufficientMove(t *testing.T) {
	b := board.New()
	from := board.Coord{0, 0}
	to := board.Coord{10, 0}
	b.SetUnit(from, 1)
	v := NewValidator(b)
	_, err := v.ValidateMove(MoveRequest{
		UnitID: 1, From: from, To: to, MovePoints: 2,
	})
	if err == nil {
		t.Fatal("expected exceeds move")
	}
}
