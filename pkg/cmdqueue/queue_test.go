package cmdqueue

import "testing"

func TestFIFOWithinPriority(t *testing.T) {
	q := New(DefaultConfig())
	q.Enqueue(TempHot, Command{ID: 1, Priority: PriorityP1, Payload: []byte("a")})
	q.Enqueue(TempHot, Command{ID: 2, Priority: PriorityP1, Payload: []byte("b")})
	c1, ok := q.Dequeue()
	if !ok || c1.ID != 1 {
		t.Fatal("fifo order")
	}
}

func TestPriorityOrder(t *testing.T) {
	q := New(DefaultConfig())
	q.Enqueue(TempHot, Command{ID: 10, Priority: PriorityP2})
	q.Enqueue(TempHot, Command{ID: 11, Priority: PriorityP0})
	c, ok := q.Dequeue()
	if !ok || c.ID != 11 {
		t.Fatal("P0 first")
	}
}

func TestTemperatureOrder(t *testing.T) {
	q := New(DefaultConfig())
	q.Enqueue(TempCold, Command{ID: 1, Priority: PriorityP0})
	q.Enqueue(TempHot, Command{ID: 2, Priority: PriorityP2})
	c, ok := q.Dequeue()
	if !ok || c.ID != 2 {
		t.Fatal("hot before cold")
	}
}

func TestCapacity(t *testing.T) {
	cfg := Config{HotCap: 1, WarmCap: 1, ColdCap: 1}
	q := New(cfg)
	if !q.Enqueue(TempHot, Command{ID: 1, Priority: PriorityP0}) {
		t.Fatal("first")
	}
	if q.Enqueue(TempHot, Command{ID: 2, Priority: PriorityP0}) {
		t.Fatal("should be full")
	}
}
