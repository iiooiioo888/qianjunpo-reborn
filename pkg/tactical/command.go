package tactical

import (
	"encoding/binary"
	"errors"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
)

// CommandKind discriminates tactical inputs.
type CommandKind uint8

const (
	KindMove   CommandKind = 1
	KindAttack CommandKind = 2
	KindPass   CommandKind = 3
)

// Command is one player input for a lockstep frame.
type Command struct {
	PlayerID uint8
	Kind     CommandKind
	UnitID   uint32
	To       board.Coord // move destination or attack target cell
}

// Encode serializes a command for replay frames.
func Encode(c Command) []byte {
	b := make([]byte, 8)
	b[0] = byte(c.Kind)
	b[1] = c.PlayerID
	binary.LittleEndian.PutUint32(b[2:6], c.UnitID)
	b[6] = byte(c.To.X)
	b[7] = byte(c.To.Y)
	return b
}

// Decode parses a replay payload.
func Decode(p []byte) (Command, error) {
	if len(p) < 8 {
		return Command{}, errors.New("tactical: short payload")
	}
	return Command{
		Kind:     CommandKind(p[0]),
		PlayerID: p[1],
		UnitID:   binary.LittleEndian.Uint32(p[2:6]),
		To:       board.Coord{X: int(p[6]), Y: int(p[7])},
	}, nil
}
