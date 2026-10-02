// Package tactical implements a Phase 2 vertical slice: 19×19 grid duel with
// authoritative validation, combat, lockstep delay, and replay hashing.
package tactical

import (
	"errors"
	"fmt"
	"sort"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/hash"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/lockstep"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/replay"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rng"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/validate"
)

const (
	PlayerCount   = 2
	maxTurnFrames = 64
)

// commandBuffer queues tactical commands with standard lockstep delay.
type commandBuffer struct {
	byExec map[uint64][]Command
}

func newCommandBuffer() *commandBuffer {
	return &commandBuffer{byExec: make(map[uint64][]Command)}
}

func (b *commandBuffer) queue(submitFrame uint64, cmd Command) {
	exec := submitFrame + uint64(lockstep.CommandDelayFrames)
	b.byExec[exec] = append(b.byExec[exec], cmd)
}

func (b *commandBuffer) pop(execFrame uint64) []Command {
	cmds := b.byExec[execFrame]
	delete(b.byExec, execFrame)
	return cmds
}

func (b *commandBuffer) pendingCount() int {
	n := 0
	for _, cmds := range b.byExec {
		n += len(cmds)
	}
	return n
}

// PendingCommandCount returns queued lockstep commands not yet executed (scheduling metadata).
func (m *Match) PendingCommandCount() int {
	if m.buffer == nil {
		return 0
	}
	return m.buffer.pendingCount()
}

// Match is an authoritative tactical battle on the 19×19 board.
type Match struct {
	Seed     uint64
	RNG      *rng.RNG
	Board    *board.Board
	Units    map[uint32]*Unit
	validate *validate.Validator
	buffer   *commandBuffer

	Frame    uint64
	Finished  bool
	Winner    uint8
	EndReason EndReason

	controlPoints []ControlPoint
	captureHold   []uint8

	combatCfg combat.Config
	counters  combat.CounterMatrix
	cardPool  combat.CardPool
	recorder  *replay.Recorder
	initial   uint64
	recording replay.Recording
	sealed    bool
}

// NewMatch creates a standard infantry vs cavalry duel with a bridged river.
func NewMatch(seed uint64) *Match {
	cfg, err := combat.DefaultConfig()
	if err != nil {
		panic(err)
	}
	return NewMatchWithConfig(seed, cfg)
}

// NewMatchWithConfig builds a duel using a combat rules snapshot (load once per match).
func NewMatchWithConfig(seed uint64, cfg combat.Config) *Match {
	b := board.New()
	for x := 0; x < board.Size; x++ {
		b.SetTerrain(board.Coord{x, 9}, board.TerrainRiver)
	}
	b.SetTerrain(board.Coord{9, 9}, board.TerrainPass)
	b.SetPassable(board.Coord{9, 9}, true)

	u0 := defaultUnit(cfg, UnitIDPlayer0, 0, combat.UnitInfantry, board.Coord{2, 8})
	u1 := defaultUnit(cfg, UnitIDPlayer1, 1, combat.UnitCavalry, board.Coord{16, 10})
	b.SetUnit(u0.Pos, u0.ID)
	b.SetUnit(u1.Pos, u1.ID)

	m := &Match{
		Seed:     seed,
		RNG:      rng.New(seed),
		Board:    b,
		Units:    map[uint32]*Unit{u0.ID: &u0, u1.ID: &u1},
		validate: validate.NewValidator(b),
		buffer:   newCommandBuffer(),
		combatCfg: cfg,
		counters:  cfg.Counters,
		cardPool:  combat.DefaultCardPool(),
		Winner:    NoWinner,
	}
	m.initial = m.StateHash()
	m.recorder = replay.NewRecorder(m.initial, m.RNG.GetState())
	return m
}

// StateHash fingerprints board + units + frame + RNG (deterministic).
func (m *Match) StateHash() uint64 {
	h := hash.New()
	h.WriteUint64(m.Frame)
	m.Board.Hash(func(v uint64) { h.WriteUint64(v) })
	ids := make([]uint32, 0, len(m.Units))
	for id := range m.Units {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		u := m.Units[id]
		if u == nil {
			continue
		}
		h.WriteUint64(uint64(u.ID))
		h.WriteUint64(uint64(u.Owner))
		h.WriteUint64(uint64(u.Stats.Type))
		h.WriteInt64(u.Stats.HP.Raw())
		h.WriteInt64(int64(u.Pos.X))
		h.WriteInt64(int64(u.Pos.Y))
	}
	st := m.RNG.GetState()
	h.WriteUint64(st.S0)
	h.WriteUint64(st.S1)
	if m.Finished {
		h.WriteUint64(uint64(m.Winner))
	}
	return h.Sum64()
}

// Submit queues a command for the current lockstep frame (validated immediately).
func (m *Match) Submit(cmd Command) error {
	if m.Finished {
		return errors.New("tactical: match finished")
	}
	if cmd.PlayerID >= PlayerCount {
		return fmt.Errorf("tactical: invalid player %d", cmd.PlayerID)
	}
	u := m.Units[cmd.UnitID]
	if u == nil || u.Owner != cmd.PlayerID {
		return validate.MoveError{Code: validate.CodeWrongStart, Message: "unit not owned by player"}
	}
	if u.Stats.HP.Raw() <= 0 {
		return errors.New("tactical: unit eliminated")
	}

	switch cmd.Kind {
	case KindPass:
	case KindMove:
		req := validate.MoveRequest{
			UnitID: cmd.UnitID, From: u.Pos, To: cmd.To, MovePoints: u.Stats.Move,
		}
		if _, err := m.validate.ValidateMove(req); err != nil {
			return err
		}
	case KindAttack:
		if err := m.validateAttack(cmd, u); err != nil {
			return err
		}
	case KindAoE:
		if err := m.validateAoE(cmd, u); err != nil {
			return err
		}
	case KindSkill:
		if err := m.validateSkill(cmd, u); err != nil {
			return err
		}
	default:
		return fmt.Errorf("tactical: unknown kind %d", cmd.Kind)
	}

	m.buffer.queue(m.Frame, cmd)
	return nil
}

func (m *Match) validateAttack(cmd Command, attacker *Unit) error {
	targetID := m.Board.GetUnit(cmd.To)
	if targetID == 0 {
		return validate.MoveError{Code: validate.CodeWrongStart, Message: "no target at cell"}
	}
	defender := m.Units[targetID]
	if defender == nil || defender.Owner == attacker.Owner {
		return validate.MoveError{Code: validate.CodeWrongStart, Message: "invalid attack target"}
	}
	if !combat.InAttackRange(attacker.Pos, cmd.To, attacker.Stats.Range) {
		return validate.MoveError{Code: validate.CodeNotAdjacent, Message: "attack out of range"}
	}
	if combat.RangedAttackNeedsLine(attacker.Stats.Range) && !combat.AttackLineClear(m.Board, attacker.Pos, cmd.To, attacker.ID, targetID) {
		return validate.MoveError{Code: validate.CodeBlocked, Message: "attack line blocked"}
	}
	return nil
}

// StepLockstep executes one lockstep turn: delayed commands, then frame++.
func (m *Match) StepLockstep() {
	if m.Finished {
		m.Frame++
		return
	}
	cmds := m.buffer.pop(m.Frame)
	sort.Slice(cmds, func(i, j int) bool {
		if cmds[i].PlayerID != cmds[j].PlayerID {
			return cmds[i].PlayerID < cmds[j].PlayerID
		}
		return cmds[i].UnitID < cmds[j].UnitID
	})
	for _, cmd := range cmds {
		m.apply(cmd)
		submit := m.Frame - uint64(lockstep.CommandDelayFrames)
		m.recorder.AddFrame(replay.FrameCommand{
			Frame:   submit,
			Player:  cmd.PlayerID,
			Payload: Encode(cmd),
		})
	}
	m.checkVictory()
	if !m.Finished {
		m.tickCapture()
	}
	if !m.Finished && m.Frame >= maxTurnFrames {
		m.decideTimeoutWinner()
	}
	m.Frame++
}

func (m *Match) apply(cmd Command) {
	u := m.Units[cmd.UnitID]
	if u == nil || u.Stats.HP.Raw() <= 0 {
		return
	}
	switch cmd.Kind {
	case KindPass:
		return
	case KindMove:
		req := validate.MoveRequest{
			UnitID: cmd.UnitID, From: u.Pos, To: cmd.To, MovePoints: u.Stats.Move,
		}
		path, err := m.validate.ValidateMove(req)
		if err != nil {
			return
		}
		if err := m.validate.ApplyMove(req, path); err != nil {
			return
		}
		u.Pos = cmd.To
	case KindAttack:
		if err := m.validateAttack(cmd, u); err != nil {
			return
		}
		targetID := m.Board.GetUnit(cmd.To)
		defender := m.Units[targetID]
		if defender == nil {
			return
		}
		m.applyStrike(u, defender)
	case KindAoE:
		if err := m.validateAoE(cmd, u); err != nil {
			return
		}
		_ = m.ApplyAoEStrike(u.ID, cmd.To, combat.DefaultAoERadius)
	case KindSkill:
		m.applySkill(cmd, u)
	}
}

// CardPool returns the skill deck gate for this match (stub).
func (m *Match) CardPool() combat.CardPool {
	if m == nil {
		return combat.CardPool{}
	}
	return m.cardPool
}

func (m *Match) applyStrike(attacker, defender *Unit) {
	atk := combat.FinalATK(attacker.Stats, defender.Stats, m.counters)
	dmg := m.combatCfg.ResolveDamage(atk, defender.Stats.BaseDEF)
	if m.combatCfg.HitEnabled() {
		dmg = m.combatCfg.ApplyHitToDamage(dmg, m.RNG.NextUint64())
	}
	if m.combatCfg.CritEnabled() && dmg.Raw() > 0 {
		dmg = m.combatCfg.ApplyCritToDamage(dmg, m.RNG.NextUint64())
	}
	defender.Stats.HP = defender.Stats.HP.Sub(dmg)
	// Deterministic battle noise for desync detection (bounded).
	_ = m.RNG.NextIntBounded(5)
	if defender.Stats.HP.Raw() <= 0 {
		m.Board.ClearUnit(defender.Pos)
	}
}

func (m *Match) checkVictory() {
	alive := make([]uint8, 0, 2)
	for _, u := range m.Units {
		if u.Stats.HP.Raw() > 0 {
			alive = append(alive, u.Owner)
		}
	}
	if len(alive) == 1 {
		m.Finished = true
		m.Winner = alive[0]
		m.EndReason = EndAnnihilation
	}
	if len(alive) == 0 {
		m.Finished = true
		m.Winner = NoWinner
		m.EndReason = EndMutualWipe
	}
}

func (m *Match) decideTimeoutWinner() {
	var best uint8 = NoWinner
	var bestHP int64 = -1
	for _, u := range m.Units {
		hp := u.Stats.HP.Raw()
		if hp > bestHP {
			bestHP = hp
			best = u.Owner
		} else if hp == bestHP && hp >= 0 {
			best = NoWinner
		}
	}
	m.Finished = true
	m.Winner = best
	m.EndReason = EndTimeout
}

// RunSchedule submits commands at frames and steps until targetFrame.
func (m *Match) RunSchedule(schedule []ScheduledCommand, targetFrame uint64) uint64 {
	idx := 0
	for m.Frame < targetFrame {
		for idx < len(schedule) && schedule[idx].SubmitFrame == m.Frame {
			_ = m.Submit(schedule[idx].Cmd)
			idx++
		}
		m.StepLockstep()
	}
	return m.FinishRecording()
}

// ScheduledCommand binds a tactical command to a lockstep submit frame.
type ScheduledCommand struct {
	SubmitFrame uint64
	Cmd         Command
}

// FinishRecording seals the replay with the terminal state hash.
func (m *Match) FinishRecording() uint64 {
	if !m.Finished {
		if m.Winner == NoWinner {
			m.checkVictory()
		}
	}
	m.Finished = true
	final := m.StateHash()
	if !m.sealed {
		m.recording = m.recorder.Finish(final)
		m.sealed = true
	}
	return m.recording.FinalHash
}

// Recording returns the sealed replay (call after FinishRecording).
func (m *Match) Recording() replay.Recording {
	return m.recording
}

// ReplayFromRecording deterministically re-simulates a stored match.
func ReplayFromRecording(rec replay.Recording) (uint64, error) {
	if err := replay.VerifyConsistency(rec); err != nil {
		return 0, err
	}
	m := NewMatch(rec.RNGSeed.S0)
	m.RNG.SetState(rec.RNGSeed)
	if m.StateHash() != rec.InitialHash {
		return 0, fmt.Errorf("tactical: initial hash mismatch: got %016x want %016x", m.StateHash(), rec.InitialHash)
	}
	sched := make([]ScheduledCommand, 0, len(rec.Frames))
	for _, fc := range rec.Frames {
		cmd, err := Decode(fc.Payload)
		if err != nil {
			return 0, err
		}
		sched = append(sched, ScheduledCommand{SubmitFrame: fc.Frame, Cmd: cmd})
	}
	final := m.RunSchedule(sched, DemoTargetFrame())
	if final != rec.FinalHash {
		return final, fmt.Errorf("tactical: final hash mismatch: got %016x want %016x", final, rec.FinalHash)
	}
	return final, nil
}

func maxFrame(sched []ScheduledCommand) uint64 {
	var max uint64
	for _, s := range sched {
		if s.SubmitFrame > max {
			max = s.SubmitFrame
		}
	}
	return max
}

// DemoSchedule returns a scripted duel (submit frames respect CommandDelayFrames).
func DemoSchedule() []ScheduledCommand {
	return []ScheduledCommand{
		{0, Command{PlayerID: 0, Kind: KindMove, UnitID: UnitIDPlayer0, To: board.Coord{5, 8}}},
		{0, Command{PlayerID: 1, Kind: KindMove, UnitID: UnitIDPlayer1, To: board.Coord{13, 10}}},
		{4, Command{PlayerID: 0, Kind: KindMove, UnitID: UnitIDPlayer0, To: board.Coord{8, 8}}},
		{4, Command{PlayerID: 1, Kind: KindMove, UnitID: UnitIDPlayer1, To: board.Coord{11, 10}}},
		{8, Command{PlayerID: 0, Kind: KindMove, UnitID: UnitIDPlayer0, To: board.Coord{9, 8}}},
		{8, Command{PlayerID: 1, Kind: KindMove, UnitID: UnitIDPlayer1, To: board.Coord{10, 10}}},
		{12, Command{PlayerID: 0, Kind: KindMove, UnitID: UnitIDPlayer0, To: board.Coord{9, 9}}},
		{12, Command{PlayerID: 1, Kind: KindPass, UnitID: UnitIDPlayer1, To: board.Coord{10, 10}}},
		{16, Command{PlayerID: 0, Kind: KindAttack, UnitID: UnitIDPlayer0, To: board.Coord{10, 10}}},
		{16, Command{PlayerID: 1, Kind: KindAttack, UnitID: UnitIDPlayer1, To: board.Coord{9, 9}}},
		{20, Command{PlayerID: 0, Kind: KindAttack, UnitID: UnitIDPlayer0, To: board.Coord{10, 10}}},
		{20, Command{PlayerID: 1, Kind: KindAttack, UnitID: UnitIDPlayer1, To: board.Coord{9, 9}}},
		{24, Command{PlayerID: 0, Kind: KindAttack, UnitID: UnitIDPlayer0, To: board.Coord{10, 10}}},
		{24, Command{PlayerID: 1, Kind: KindAttack, UnitID: UnitIDPlayer1, To: board.Coord{9, 9}}},
		{28, Command{PlayerID: 0, Kind: KindAttack, UnitID: UnitIDPlayer0, To: board.Coord{10, 10}}},
		{28, Command{PlayerID: 1, Kind: KindPass, UnitID: UnitIDPlayer1, To: board.Coord{10, 10}}},
	}
}

// DemoTargetFrame is the lockstep frame count for DemoSchedule (includes delay tail).
func DemoTargetFrame() uint64 {
	return maxFrame(DemoSchedule()) + uint64(lockstep.CommandDelayFrames) + 3
}
