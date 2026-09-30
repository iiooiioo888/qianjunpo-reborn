package agones

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Coordinator wires Allocate → Ready → Shutdown with timeouts and metrics.
type Coordinator struct {
	cfg Config

	alloc Allocator
	gs    GameServerSide

	mu             sync.RWMutex
	phase          Phase
	identity       RoomIdentity
	lastErr        string
	lastTransition time.Time

	mock *MockBackend
}

// NewCoordinator builds backends from configuration.
func NewCoordinator(cfg Config) (*Coordinator, error) {
	c := &Coordinator{cfg: cfg, phase: PhaseUninitialized}

	switch cfg.AllocatorBackend {
	case AllocatorMock, "":
		m := NewMockBackend(cfg.GameServerPort)
		c.mock = m
		c.alloc = m
	case AllocatorHTTP:
		c.alloc = NewHTTPAllocator(cfg)
	default:
		return nil, fmt.Errorf("agones: unknown allocator %q", cfg.AllocatorBackend)
	}

	switch cfg.Backend {
	case BackendMock, "":
		if c.mock == nil {
			c.mock = NewMockBackend(cfg.GameServerPort)
		}
		c.gs = c.mock
	case BackendSidecar:
		c.gs = NewSidecarClient(cfg.SidecarBaseURL)
	default:
		return nil, fmt.Errorf("agones: unknown backend %q", cfg.Backend)
	}
	return c, nil
}

// NewFromEnv loads Config and constructs a Coordinator.
func NewFromEnv() (*Coordinator, error) {
	return NewCoordinator(LoadConfig())
}

// AllocateRoom runs the control-plane allocation step (mock or HTTP).
func (c *Coordinator) AllocateRoom(ctx context.Context, req AllocateRequest) (RoomIdentity, error) {
	ctx, cancel := c.withTimeout(ctx, c.cfg.AllocateTimeout)
	defer cancel()

	ident, err := c.alloc.Allocate(ctx, req)
	if err != nil {
		c.recordError("allocate", err, ctx.Err() == context.DeadlineExceeded)
		observeOp("allocate", "error")
		return RoomIdentity{}, &OperationError{Operation: "allocate", Phase: c.phase, Err: err, Timeout: ctx.Err() == context.DeadlineExceeded}
	}
	c.mu.Lock()
	c.identity = ident
	c.phase = PhaseAllocated
	c.lastErr = ""
	c.lastTransition = time.Now().UTC()
	c.mu.Unlock()
	observeOp("allocate", "ok")
	log.Printf("agones/lifecycle: allocated room name=%s addr=%s:%d room_id=%s", ident.Name, ident.Address, ident.Port, ident.RoomID)
	return ident, nil
}

// BootstrapGameServer marks the running Roma pod Ready after listeners are up.
func (c *Coordinator) BootstrapGameServer(ctx context.Context) error {
	ctx, cancel := c.withTimeout(ctx, c.cfg.ReadyTimeout)
	defer cancel()

	if ident, err := c.gs.Identity(ctx); err == nil && ident.Name != "" {
		c.mu.Lock()
		c.identity = ident
		c.mu.Unlock()
	}

	err := c.gs.Ready(ctx)
	if err != nil {
		c.recordError("ready", err, ctx.Err() == context.DeadlineExceeded)
		observeOp("ready", "error")
		return &OperationError{Operation: "ready", Phase: c.phase, Err: err, Timeout: ctx.Err() == context.DeadlineExceeded}
	}
	c.mu.Lock()
	c.phase = PhaseReady
	c.lastErr = ""
	c.lastTransition = time.Now().UTC()
	c.mu.Unlock()
	observeOp("ready", "ok")
	log.Printf("agones/lifecycle: ready backend=%s", c.cfg.Backend)
	return nil
}

// ShutdownGameServer signals Agones to recycle the pod (or mock equivalent).
func (c *Coordinator) ShutdownGameServer(ctx context.Context) error {
	ctx, cancel := c.withTimeout(ctx, c.cfg.ShutdownTimeout)
	defer cancel()

	c.mu.Lock()
	c.phase = PhaseShuttingDown
	c.lastTransition = time.Now().UTC()
	c.mu.Unlock()

	err := c.gs.Shutdown(ctx)
	if err != nil {
		c.recordError("shutdown", err, ctx.Err() == context.DeadlineExceeded)
		observeOp("shutdown", "error")
		return &OperationError{Operation: "shutdown", Phase: c.phase, Err: err, Timeout: ctx.Err() == context.DeadlineExceeded}
	}
	c.mu.Lock()
	c.phase = PhaseShutdown
	c.lastErr = ""
	c.lastTransition = time.Now().UTC()
	c.mu.Unlock()
	observeOp("shutdown", "ok")
	log.Printf("agones/lifecycle: shutdown complete backend=%s", c.cfg.Backend)
	return nil
}

// RunHealthLoop pings sidecar/mock health until ctx is cancelled.
func (c *Coordinator) RunHealthLoop(ctx context.Context) {
	tick := time.NewTicker(c.cfg.HealthInterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			hctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			err := c.gs.Health(hctx)
			cancel()
			if err != nil {
				observeOp("health", "error")
				log.Printf("agones/lifecycle: health ping failed: %v", err)
			} else {
				observeOp("health", "ok")
			}
		}
	}
}

// Status returns the latest lifecycle snapshot for HTTP and ops.
func (c *Coordinator) Status(ctx context.Context) StatusSnapshot {
	c.mu.RLock()
	coordPhase := c.phase
	coordErr := c.lastErr
	coordIdent := c.identity
	coordTransition := c.lastTransition
	c.mu.RUnlock()

	if c.cfg.Backend == BackendSidecar {
		cli := NewSidecarClient(c.cfg.SidecarBaseURL)
		if snap, err := cli.SidecarStatusSnapshot(ctx); err == nil {
			return snap
		}
	}
	if c.mock != nil {
		snap := c.mock.MockSnapshot()
		if coordErr != "" {
			snap.Phase = coordPhase
			snap.LastError = coordErr
			snap.LastTransitionUTC = coordTransition
			snap.Ready = coordPhase == PhaseReady
		}
		if coordIdent.Name != "" {
			snap.Identity = coordIdent
		}
		return snap
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return StatusSnapshot{
		Backend:           c.cfg.Backend,
		AllocatorBackend:  c.cfg.AllocatorBackend,
		Phase:             c.phase,
		Identity:          c.identity,
		LastError:         c.lastErr,
		LastTransitionUTC: c.lastTransition,
		Ready:             c.phase == PhaseReady,
	}
}

func (c *Coordinator) withTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, d)
}

func (c *Coordinator) recordError(op string, err error, timeout bool) {
	c.mu.Lock()
	c.phase = PhaseError
	c.lastErr = err.Error()
	c.lastTransition = time.Now().UTC()
	c.mu.Unlock()
	if timeout {
		log.Printf("agones/lifecycle: %s timeout: %v", op, err)
	} else {
		log.Printf("agones/lifecycle: %s failed: %v", op, err)
	}
}

// MockBackendForTest exposes the mock implementation for tests.
func (c *Coordinator) MockBackendForTest() *MockBackend { return c.mock }
