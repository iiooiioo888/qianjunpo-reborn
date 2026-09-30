package sim

import (
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/lockstep"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rng"
)

// ScheduledCommand records when a command was submitted in lockstep frames.
type ScheduledCommand struct {
	SubmitFrame uint64
	Cmd         lockstep.CommandPacket
}

// Engine runs lockstep turns with delayed commands and sub-frame simulation.
type Engine struct {
	Seed    uint64
	RNG     *rng.RNG
	State   GameState
	Buffer  *lockstep.CommandBuffer
	History []uint64
}

// NewEngine creates an engine with deterministic RNG seed.
func NewEngine(seed uint64) *Engine {
	gs := NewGameState()
	r := rng.New(seed)
	return &Engine{
		Seed:   seed,
		RNG:    r,
		State:  gs,
		Buffer: lockstep.NewCommandBuffer(),
	}
}

// SubmitCommand queues input for the current lockstep frame.
func (e *Engine) SubmitCommand(cmd lockstep.CommandPacket) {
	e.Buffer.QueueSubmission(e.State.LockstepFrame, cmd)
}

// CurrentHash returns the digest after the last completed logic step.
func (e *Engine) CurrentHash() uint64 {
	return Hash(e.State, e.RNG.GetState())
}

// StepLockstep executes one lockstep turn (6 sub-frames) with optimistic fill.
func (e *Engine) StepLockstep() uint64 {
	frame := e.State.LockstepFrame
	raw := e.Buffer.PopExecutable(frame)
	cmds := lockstep.FillOptimistic(raw, PlayerCount)

	var lastHash uint64
	for sub := 0; sub < lockstep.GameTurnFrames; sub++ {
		e.State.SubFrame = uint8(sub)
		applySubFrame(&e.State, cmds, e.RNG)
		lastHash = Hash(e.State, e.RNG.GetState())
		e.History = append(e.History, lastHash)
	}
	e.State.LockstepFrame++
	e.State.SubFrame = 0
	return lastHash
}

// RunWithSchedule submits commands at their frames and advances until targetFrame lockstep turns complete.
func (e *Engine) RunWithSchedule(schedule []ScheduledCommand, targetFrame uint64) uint64 {
	idx := 0
	var h uint64
	for e.State.LockstepFrame < targetFrame {
		for idx < len(schedule) && schedule[idx].SubmitFrame == e.State.LockstepFrame {
			e.Buffer.QueueSubmission(schedule[idx].SubmitFrame, schedule[idx].Cmd)
			idx++
		}
		h = e.StepLockstep()
	}
	return h
}

// RestoreSnapshot resets engine state from a snapshot (rollback entry point).
func (e *Engine) RestoreSnapshot(s Snapshot) {
	e.State = cloneGameState(s.GameState)
	e.RNG.SetState(s.RNGState)
	e.Buffer = lockstep.NewCommandBuffer()
	e.History = nil
}

// ResimulateFromSnapshot restores and replays schedule from snapshot frame to targetFrame.
func (e *Engine) ResimulateFromSnapshot(s Snapshot, schedule []ScheduledCommand, targetFrame uint64) uint64 {
	e.RestoreSnapshot(s)
	start := s.GameState.LockstepFrame
	var future []ScheduledCommand
	for _, sc := range schedule {
		execFrame := sc.SubmitFrame + uint64(lockstep.CommandDelayFrames)
		if execFrame < start {
			continue
		}
		if sc.SubmitFrame < start {
			e.Buffer.QueueSubmission(sc.SubmitFrame, sc.Cmd)
			continue
		}
		future = append(future, sc)
	}
	return e.RunWithSchedule(future, targetFrame)
}
