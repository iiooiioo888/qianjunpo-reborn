package agones

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// MockBackend implements both Allocator and GameServerSide for CI and local pre-prod.
type MockBackend struct {
	mu sync.RWMutex

	seq            atomic.Uint64
	phase          Phase
	identity       RoomIdentity
	lastErr        string
	lastTransition time.Time
	readyDelay     time.Duration
	failReady      bool
	failShutdown   bool
	port           int
}

// NewMockBackend returns a mock with optional Roma gRPC port for allocate responses.
func NewMockBackend(port int) *MockBackend {
	if port <= 0 {
		port = 9092
	}
	return &MockBackend{
		phase: PhaseUninitialized,
		port:  port,
	}
}

// SetReadyDelay simulates slow Ready for timeout tests.
func (b *MockBackend) SetReadyDelay(d time.Duration) { b.readyDelay = d }

// SetFailReady forces the next Ready call to fail.
func (b *MockBackend) SetFailReady(v bool) { b.failReady = v }

// SetFailShutdown forces the next Shutdown call to fail.
func (b *MockBackend) SetFailShutdown(v bool) { b.failShutdown = v }

func (b *MockBackend) Allocate(_ context.Context, req AllocateRequest) (RoomIdentity, error) {
	id := b.seq.Add(1)
	name := fmt.Sprintf("mock-gs-%d", id)
	roomID := fmt.Sprintf("room-%s-%d", req.ZoneID, req.Shard)
	ident := RoomIdentity{
		Name:    name,
		Address: "127.0.0.1",
		Port:    b.port,
		RoomID:  roomID,
	}
	b.mu.Lock()
	b.identity = ident
	b.phase = PhaseAllocated
	b.lastErr = ""
	b.lastTransition = time.Now().UTC()
	b.mu.Unlock()
	return ident, nil
}

func (b *MockBackend) Identity(_ context.Context) (RoomIdentity, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.identity, nil
}

func (b *MockBackend) Ready(ctx context.Context) error {
	if b.readyDelay > 0 {
		select {
		case <-ctx.Done():
			b.mu.Lock()
			b.phase = PhaseError
			b.lastErr = ctx.Err().Error()
			b.lastTransition = time.Now().UTC()
			b.mu.Unlock()
			return ctx.Err()
		case <-time.After(b.readyDelay):
		}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.failReady {
		b.phase = PhaseError
		b.lastErr = "mock ready failure"
		b.lastTransition = time.Now().UTC()
		return fmt.Errorf("mock: ready rejected")
	}
	if b.identity.Name == "" {
		b.identity = RoomIdentity{Name: "mock-local", Address: "127.0.0.1", Port: b.port}
	}
	b.phase = PhaseReady
	b.lastErr = ""
	b.lastTransition = time.Now().UTC()
	return nil
}

func (b *MockBackend) Shutdown(_ context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.failShutdown {
		b.phase = PhaseError
		b.lastErr = "mock shutdown failure"
		b.lastTransition = time.Now().UTC()
		return fmt.Errorf("mock: shutdown rejected")
	}
	b.phase = PhaseShutdown
	b.lastErr = ""
	b.lastTransition = time.Now().UTC()
	return nil
}

func (b *MockBackend) Health(_ context.Context) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.phase == PhaseError {
		return fmt.Errorf("mock: unhealthy (%s)", b.lastErr)
	}
	return nil
}

// MockSnapshot exports internal phase for status HTTP / tests.
func (b *MockBackend) MockSnapshot() StatusSnapshot {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return StatusSnapshot{
		Backend:           BackendMock,
		AllocatorBackend:  AllocatorMock,
		Phase:             b.phase,
		Identity:          b.identity,
		LastError:         b.lastErr,
		LastTransitionUTC: b.lastTransition,
		Ready:             b.phase == PhaseReady,
	}
}
