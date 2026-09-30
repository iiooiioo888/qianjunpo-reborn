package cmdqueue

import (
	"testing"
	"time"
)

func TestFIFOWithinPriority(t *testing.T) {
	q := New(Config{HotCap: 8, WarmCap: 8, ColdTTL: time.Second})
	q.Enqueue(TempHot, Command{ID: 1, Priority: PriorityP1, Payload: []byte("a")})
	q.Enqueue(TempHot, Command{ID: 2, Priority: PriorityP1, Payload: []byte("b")})
	c1, ok := q.Dequeue()
	if !ok || c1.ID != 1 {
		t.Fatal("fifo order")
	}
}

func TestPriorityOrder(t *testing.T) {
	q := New(Config{HotCap: 8, WarmCap: 8, ColdTTL: time.Second})
	q.Enqueue(TempHot, Command{ID: 10, Priority: PriorityP2})
	q.Enqueue(TempHot, Command{ID: 11, Priority: PriorityP0})
	c, ok := q.Dequeue()
	if !ok || c.ID != 11 {
		t.Fatal("P0 first")
	}
}

func TestTemperatureOrder(t *testing.T) {
	q := New(Config{HotCap: 8, WarmCap: 8, ColdTTL: time.Second})
	q.Enqueue(TempCold, Command{ID: 1, Priority: PriorityP0})
	q.Enqueue(TempHot, Command{ID: 2, Priority: PriorityP2})
	c, ok := q.Dequeue()
	if !ok || c.ID != 2 {
		t.Fatal("hot before cold")
	}
}

func TestHotTierCapacity(t *testing.T) {
	cfg := Config{HotCap: 1, WarmCap: 8, ColdTTL: time.Second}
	q := New(cfg)
	if !q.Enqueue(TempHot, Command{ID: 1, Priority: PriorityP0}) {
		t.Fatal("first")
	}
	if q.Enqueue(TempHot, Command{ID: 2, Priority: PriorityP0}) {
		t.Fatal("hot tier should be full")
	}
	if q.HotLen() != 1 {
		t.Fatalf("HotLen=%d want 1", q.HotLen())
	}
}

func TestDefaultConfigCaps(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.HotCap != DefaultHotCap || cfg.WarmCap != DefaultWarmCap || cfg.ColdTTL != DefaultColdTTL {
		t.Fatalf("defaults: %+v", cfg)
	}
}

func TestTryEnqueueSpillHotToWarmToCold(t *testing.T) {
	q := New(Config{HotCap: 1, WarmCap: 1, ColdTTL: time.Minute})
	out := q.TryEnqueue(Command{ID: 1, Priority: PriorityP1})
	if !out.OK || out.StoredAt != TempHot {
		t.Fatalf("first hot: %+v", out)
	}
	out = q.TryEnqueue(Command{ID: 2, Priority: PriorityP1})
	if !out.OK || out.StoredAt != TempWarm {
		t.Fatalf("spill warm: %+v", out)
	}
	out = q.TryEnqueue(Command{ID: 3, Priority: PriorityP1})
	if !out.OK || out.StoredAt != TempCold {
		t.Fatalf("spill cold: %+v", out)
	}
	if q.Len() != 3 {
		t.Fatalf("Len=%d", q.Len())
	}
}

func TestTryEnqueueDropsP2WhenHotAndWarmFull(t *testing.T) {
	q := New(Config{HotCap: 1, WarmCap: 1, ColdTTL: time.Minute})
	if !q.TryEnqueue(Command{ID: 1, Priority: PriorityP1}).OK {
		t.Fatal("fill hot")
	}
	if !q.TryEnqueue(Command{ID: 2, Priority: PriorityP1}).OK {
		t.Fatal("fill warm")
	}
	out := q.TryEnqueue(Command{ID: 99, Priority: PriorityP2})
	if out.OK || !out.Dropped {
		t.Fatalf("P2 should drop when hot+warm full: %+v", out)
	}
}

func TestTryEnqueueP1UsesColdWhenHotWarmFull(t *testing.T) {
	q := New(Config{HotCap: 1, WarmCap: 1, ColdTTL: time.Minute})
	_ = q.TryEnqueue(Command{ID: 1, Priority: PriorityP0})
	_ = q.TryEnqueue(Command{ID: 2, Priority: PriorityP0})
	out := q.TryEnqueue(Command{ID: 3, Priority: PriorityP1})
	if !out.OK || out.StoredAt != TempCold {
		t.Fatalf("P1 cold spill: %+v", out)
	}
}


func TestTryEnqueueP0NeverDroppedUsesCold(t *testing.T) {
	q := New(Config{HotCap: 0, WarmCap: 0, ColdTTL: time.Minute})
	out := q.TryEnqueue(Command{ID: 42, Priority: PriorityP0})
	if !out.OK || out.Dropped || out.StoredAt != TempCold {
		t.Fatalf("P0 cold: %+v", out)
	}
	c, ok := q.Dequeue()
	if !ok || c.ID != 42 {
		t.Fatal("dequeue P0")
	}
}

func TestColdTTLExpiry(t *testing.T) {
	start := time.Unix(1_700_000_000, 0)
	now := start
	q := New(Config{HotCap: 4, WarmCap: 4, ColdTTL: 5 * time.Second})
	q.SetClock(func() time.Time { return now })
	q.Enqueue(TempCold, Command{ID: 1, Priority: PriorityP2})
	q.Enqueue(TempCold, Command{ID: 2, Priority: PriorityP2})
	if q.ColdLen() != 2 {
		t.Fatalf("ColdLen=%d", q.ColdLen())
	}
	now = start.Add(6 * time.Second)
	q.PurgeColdExpired()
	if q.ColdLen() != 0 {
		t.Fatalf("expected expired cold empty, ColdLen=%d", q.ColdLen())
	}
	_, ok := q.Dequeue()
	if ok {
		t.Fatal("nothing left after ttl")
	}
}

func TestColdTTLFIFOBeforeExpiry(t *testing.T) {
	start := time.Unix(1_700_000_000, 0)
	q := New(Config{HotCap: 4, WarmCap: 4, ColdTTL: 10 * time.Second})
	q.SetClock(func() time.Time { return start })
	q.Enqueue(TempCold, Command{ID: 10, Priority: PriorityP1})
	q.Enqueue(TempCold, Command{ID: 11, Priority: PriorityP1})
	c, ok := q.Dequeue()
	if !ok || c.ID != 10 {
		t.Fatal("cold fifo")
	}
}

func TestP0StaleHookOnColdExpiry(t *testing.T) {
	start := time.Unix(1_700_000_000, 0)
	now := start
	var staleID uint64
	q := New(Config{HotCap: 1, WarmCap: 1, ColdTTL: 5 * time.Second})
	q.SetClock(func() time.Time { return now })
	q.SetP0StaleHook(func(c Command) { staleID = c.ID })
	q.Enqueue(TempCold, Command{ID: 100, Priority: PriorityP0})
	now = start.Add(6 * time.Second)
	q.PurgeColdExpired()
	if staleID != 100 {
		t.Fatalf("stale hook id=%d", staleID)
	}
}

func TestP2ColdExpirySilent(t *testing.T) {
	start := time.Unix(1_700_000_000, 0)
	now := start
	hookCalled := false
	q := New(Config{HotCap: 1, WarmCap: 1, ColdTTL: 5 * time.Second})
	q.SetClock(func() time.Time { return now })
	q.SetP0StaleHook(func(Command) { hookCalled = true })
	q.Enqueue(TempCold, Command{ID: 1, Priority: PriorityP2})
	now = start.Add(6 * time.Second)
	q.PurgeColdExpired()
	if hookCalled {
		t.Fatal("P2 expiry should not invoke P0 hook")
	}
}

func TestNextID(t *testing.T) {
	q := New(DefaultConfig())
	if q.NextID() != 1 || q.NextID() != 2 {
		t.Fatal("monotonic ids")
	}
}
