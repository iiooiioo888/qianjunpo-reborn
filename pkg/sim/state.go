package sim

import (
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/hash"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/lockstep"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rng"
)

const PlayerCount = 2

// PlayerState is minimal demo state using fixed-point positions.
type PlayerState struct {
	Position fixed.Vector2
	Ticks    uint64
}

// GameState is the authoritative simulation state hashed each logic frame.
type GameState struct {
	LockstepFrame uint64
	SubFrame      uint8
	Players       []PlayerState
}

// NewGameState creates initial player positions (FromFloat only at init).
func NewGameState() GameState {
	return GameState{
		Players: []PlayerState{
			{Position: fixed.NewVector2(fixed.FromInt(0), fixed.FromInt(0))},
			{Position: fixed.NewVector2(fixed.FromInt(10), fixed.FromInt(10))},
		},
	}
}

// Hash computes a deterministic FNV-1a digest of game + RNG state.
func Hash(gs GameState, rngState rng.State) uint64 {
	h := hash.New()
	h.WriteUint64(gs.LockstepFrame)
	h.WriteUint64(uint64(gs.SubFrame))
	h.WriteUint64(rngState.S0)
	h.WriteUint64(rngState.S1)
	for _, p := range gs.Players {
		h.WriteInt64(p.Position.X.Raw())
		h.WriteInt64(p.Position.Y.Raw())
		h.WriteUint64(p.Ticks)
	}
	return h.Sum64()
}

// applyCommands runs one logic sub-frame for the given lockstep commands.
func applySubFrame(gs *GameState, cmds []lockstep.CommandPacket, rngInst *rng.RNG) {
	step := fixed.FromInt(1).Div(fixed.FromInt(int64(lockstep.GameTurnFrames)))
	for _, cmd := range cmds {
		pid := int(cmd.PlayerID)
		if pid < 0 || pid >= len(gs.Players) {
			continue
		}
		dx := fixed.FromInt(int64(cmd.MoveX)).Mul(step)
		dy := fixed.FromInt(int64(cmd.MoveY)).Mul(step)
		p := &gs.Players[pid]
		p.Position = p.Position.Add(fixed.NewVector2(dx, dy))
		p.Ticks++
		// Small RNG-driven jitter for desync testing (bounded, no floats).
		if cmd.MoveX != 0 || cmd.MoveY != 0 {
			j := rngInst.NextIntBounded(3)
			p.Ticks += uint64(j)
		}
	}
}
