package sim

import "github.com/iiooiioo888/qianjunpo-reborn/pkg/rng"

// Snapshot captures restorable simulation state at a logic frame boundary.
type Snapshot struct {
	FrameNumber uint64
	SubFrame    uint8
	RNGState    rng.State
	GameState   GameState
	StateHash   uint64
}

// TakeSnapshot records the current engine state.
func TakeSnapshot(gs GameState, rngInst *rng.RNG) Snapshot {
	st := rngInst.GetState()
	return Snapshot{
		FrameNumber: gs.LockstepFrame,
		SubFrame:    gs.SubFrame,
		RNGState:    st,
		GameState:   cloneGameState(gs),
		StateHash:   Hash(gs, st),
	}
}

func cloneGameState(gs GameState) GameState {
	out := GameState{
		LockstepFrame: gs.LockstepFrame,
		SubFrame:      gs.SubFrame,
		Players:       make([]PlayerState, len(gs.Players)),
	}
	for i, p := range gs.Players {
		out.Players[i] = p
	}
	return out
}
