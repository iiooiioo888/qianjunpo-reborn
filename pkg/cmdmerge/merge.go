// Package cmdmerge coalesces player commands under overload while preserving determinism.
package cmdmerge

import (
	"encoding/binary"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/cmdqueue"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timedilation"
)

// CommandKind discriminates mergeable payloads.
type CommandKind uint8

const (
	KindMove  CommandKind = 1
	KindBuild CommandKind = 2
	KindOther CommandKind = 255
)

const P2BatchWindow = 500 * time.Millisecond

// MergePolicy decides when merging is active.
type MergePolicy struct {
	// Activate when hot queue is full and time flow at or below this rate.
	RateThreshold timedilation.Rate
	HotCap        int
}

// DefaultMergePolicy matches Phase 3 brief (0.1x rate, default hot cap).
func DefaultMergePolicy() MergePolicy {
	return MergePolicy{RateThreshold: timedilation.MinRate, HotCap: 64}
}

// ShouldActivate returns true when overload merging should run.
func (p MergePolicy) ShouldActivate(hotLen int, rate timedilation.Rate) bool {
	return hotLen >= p.HotCap && rate <= p.RateThreshold
}

// MovePayload encodes a grid move target (deterministic int8 deltas).
type MovePayload struct {
	DX, DY int8
}

// BuildPayload encodes build tile.
type BuildPayload struct {
	X, Y int16
}

func encodeMove(m MovePayload) []byte {
	b := make([]byte, 3)
	b[0] = byte(KindMove)
	b[1] = byte(m.DX)
	b[2] = byte(m.DY)
	return b
}

func encodeBuild(b BuildPayload) []byte {
	out := make([]byte, 5)
	out[0] = byte(KindBuild)
	binary.LittleEndian.PutUint16(out[1:3], uint16(b.X))
	binary.LittleEndian.PutUint16(out[3:5], uint16(b.Y))
	return out
}

// ParseKind reads the leading kind byte.
func ParseKind(payload []byte) CommandKind {
	if len(payload) == 0 {
		return KindOther
	}
	return CommandKind(payload[0])
}

// Merger statefully merges consecutive move/build commands.
type Merger struct {
	policy        MergePolicy
	lastMove      *MovePayload
	lastBuild     *BuildPayload
	p2Batch       []cmdqueue.Command
	p2WindowStart time.Time
	clock         func() time.Time
}

// NewMerger creates a merger with optional clock (defaults to time.Now).
func NewMerger(policy MergePolicy, clock func() time.Time) *Merger {
	if clock == nil {
		clock = time.Now
	}
	return &Merger{policy: policy, clock: clock}
}

// Push ingests a command; returns zero or one merged output when policy active.
func (m *Merger) Push(c cmdqueue.Command, hotLen int, rate timedilation.Rate) (out []cmdqueue.Command, ok bool) {
	active := m.policy.ShouldActivate(hotLen, rate)
	kind := ParseKind(c.Payload)

	if c.Priority == cmdqueue.PriorityP2 && active {
		now := m.clock()
		if m.p2WindowStart.IsZero() || now.Sub(m.p2WindowStart) > P2BatchWindow {
			if len(m.p2Batch) > 0 {
				out = append(out, m.flushP2()...)
			}
			m.p2WindowStart = now
			m.p2Batch = nil
		}
		m.p2Batch = append(m.p2Batch, c)
		return nil, true
	}

	if !active {
		return []cmdqueue.Command{c}, true
	}

	switch kind {
	case KindMove:
		mv := MovePayload{DX: int8(c.Payload[1]), DY: int8(c.Payload[2])}
		m.lastMove = &mv
		return nil, true
	case KindBuild:
		b := BuildPayload{
			X: int16(binary.LittleEndian.Uint16(c.Payload[1:3])),
			Y: int16(binary.LittleEndian.Uint16(c.Payload[3:5])),
		}
		m.lastBuild = &b
		return nil, true
	default:
		var flush []cmdqueue.Command
		flush = append(flush, m.flushPending()...)
		flush = append(flush, c)
		return flush, true
	}
}

// Flush emits pending merged move/build and P2 batch.
func (m *Merger) Flush() []cmdqueue.Command {
	out := m.flushPending()
	out = append(out, m.flushP2()...)
	return out
}

func (m *Merger) flushPending() []cmdqueue.Command {
	var out []cmdqueue.Command
	if m.lastMove != nil {
		out = append(out, cmdqueue.Command{Priority: cmdqueue.PriorityP1, Payload: encodeMove(*m.lastMove)})
		m.lastMove = nil
	}
	if m.lastBuild != nil {
		out = append(out, cmdqueue.Command{Priority: cmdqueue.PriorityP1, Payload: encodeBuild(*m.lastBuild)})
		m.lastBuild = nil
	}
	return out
}

func (m *Merger) flushP2() []cmdqueue.Command {
	if len(m.p2Batch) == 0 {
		return nil
	}
	merged := m.p2Batch[len(m.p2Batch)-1]
	m.p2Batch = nil
	m.p2WindowStart = time.Time{}
	return []cmdqueue.Command{merged}
}

// NewMoveCommand helper for tests and AI tactical layer.
func NewMoveCommand(id uint64, dx, dy int8) cmdqueue.Command {
	return cmdqueue.Command{
		ID:       id,
		Priority: cmdqueue.PriorityP1,
		Payload:  encodeMove(MovePayload{DX: dx, DY: dy}),
	}
}
