// Package backend implements pluggable edge inference backends (mock, Ollama).
package backend

import (
	"context"
	"time"
)

// InferRequest is the edge /v1/infer JSON body.
type InferRequest struct {
	Prompt string
}

// InferResponse matches the HTTP API contract.
type InferResponse struct {
	Text      string
	Model     string
	LatencyMs int64
}

// Health reports runtime backend status for /health.
type Health struct {
	Status  string `json:"status"`
	Model   string `json:"model"`
	Backend string `json:"backend"`
	Ready   bool   `json:"ready"`
	Detail  string `json:"detail,omitempty"`
}

// Backend runs inference for a single prompt.
type Backend interface {
	Name() string
	Model() string
	Infer(ctx context.Context, req InferRequest) (InferResponse, error)
	Health(ctx context.Context) Health
}

// Clock abstracts time for tests.
type Clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}

type realClock struct{}

func (realClock) Now() time.Time                  { return time.Now() }
func (realClock) Since(t time.Time) time.Duration { return time.Since(t) }
