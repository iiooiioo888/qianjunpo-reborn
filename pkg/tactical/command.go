package tactical

import (
	"encoding/binary"
	"errors"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
)

// CommandKind discriminates tactical inputs.
type CommandKind uint8

const (
	KindMove   CommandKind = 1
	KindAttack CommandKind = 2
	KindPass   CommandKind = 3
	KindAoE    CommandKind = 4
	KindSkill  CommandKind = 5
)

// Command is one player input for a lockstep frame.
type Command struct {
	PlayerID uint8
	Kind     CommandKind
	UnitID   uint32
	To       board.Coord // move destination, attack target cell, AoE center, or skill target
	SkillID  combat.SkillID // KindSkill only; 0 for other kinds
}

const commandPayloadV1Len = 8
const commandPayloadV2Len = 10

// Encode serializes a command for replay frames (v2 adds skill_id for KindSkill).
func Encode(c Command) []byte {
	n := commandPayloadV1Len
	if c.Kind == KindSkill || c.SkillID != 0 {
		n = commandPayloadV2Len
	}
	b := make([]byte, n)
	b[0] = byte(c.Kind)
	b[1] = c.PlayerID
	binary.LittleEndian.PutUint32(b[2:6], c.UnitID)
	b[6] = byte(c.To.X)
	b[7] = byte(c.To.Y)
	if n >= commandPayloadV2Len {
		binary.LittleEndian.PutUint16(b[8:10], uint16(c.SkillID))
	}
	return b
}

// Decode parses a replay payload.
func Decode(p []byte) (Command, error) {
	if len(p) < commandPayloadV1Len {
		return Command{}, errors.New("tactical: short payload")
	}
	cmd := Command{
		Kind:     CommandKind(p[0]),
		PlayerID: p[1],
		UnitID:   binary.LittleEndian.Uint32(p[2:6]),
		To:       board.Coord{X: int(p[6]), Y: int(p[7])},
	}
	if len(p) >= commandPayloadV2Len {
		cmd.SkillID = combat.SkillID(binary.LittleEndian.Uint16(p[8:10]))
	}
	return cmd, nil
}
