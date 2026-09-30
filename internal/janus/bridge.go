package janus

import (
	"context"
	"sync"
	"time"
)

// RateLimiter is a placeholder token bucket (skeleton).
type RateLimiter struct {
	mu       sync.Mutex
	tokens   int
	capacity int
}

// NewRateLimiter creates a limiter with burst capacity.
func NewRateLimiter(capacity int) *RateLimiter {
	if capacity <= 0 {
		capacity = 100
	}
	return &RateLimiter{tokens: capacity, capacity: capacity}
}

// Allow consumes one token if available.
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.tokens <= 0 {
		return false
	}
	r.tokens--
	return true
}

// AuthHook validates access tokens via Lares.
type AuthHook interface {
	ValidateAccess(ctx context.Context, token string) (playerID uint64, ok bool)
}

// StaticAuth accepts non-empty tokens in dev.
type StaticAuth struct{}

func (StaticAuth) ValidateAccess(_ context.Context, token string) (uint64, bool) {
	if token == "" {
		return 0, false
	}
	return 1, true
}

// Session bridges TCP client I/O to internal gRPC (skeleton state).
type Session struct {
	ID        string
	PlayerID  uint64
	ZoneID    string
	CreatedAt time.Time
}
