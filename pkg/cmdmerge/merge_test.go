package cmdmerge

import (
	"testing"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/cmdqueue"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timedilation"
)

func TestMergeConsecutiveMoves(t *testing.T) {
	m := NewMerger(DefaultMergePolicy(), func() time.Time { return time.Unix(0, 0) })
	hot := 64
	var rate timedilation.Rate = timedilation.MinRate
	c1 := NewMoveCommand(1, 1, 0)
	c2 := NewMoveCommand(2, 0, 1)
	if _, ok := m.Push(c1, hot, rate); !ok {
		t.Fatal("push1")
	}
	if _, ok := m.Push(c2, hot, rate); !ok {
		t.Fatal("push2")
	}
	out := m.Flush()
	if len(out) != 1 {
		t.Fatalf("want 1 merged move, got %d", len(out))
	}
	if out[0].Payload[1] != 0 || out[0].Payload[2] != 1 {
		t.Fatalf("last move wins: %v %v", out[0].Payload[1], out[0].Payload[2])
	}
}

func TestInactiveWhenHealthy(t *testing.T) {
	m := NewMerger(DefaultMergePolicy(), time.Now)
	c := NewMoveCommand(1, 1, 1)
	out, ok := m.Push(c, 0, timedilation.MaxRate)
	if !ok || len(out) != 1 {
		t.Fatal("should pass through")
	}
}

func TestP2Batching(t *testing.T) {
	start := time.Unix(0, 0)
	now := start
	m := NewMerger(DefaultMergePolicy(), func() time.Time { return now })
	hot := 64
	var rate timedilation.Rate = timedilation.MinRate
	for i := 0; i < 3; i++ {
		c := cmdqueue.Command{ID: uint64(i), Priority: cmdqueue.PriorityP2, Payload: []byte{99}}
		m.Push(c, hot, rate)
		now = now.Add(100 * time.Millisecond)
	}
	out := m.Flush()
	if len(out) != 1 || out[0].ID != 2 {
		t.Fatalf("batch should keep last P2: %+v", out)
	}
}
