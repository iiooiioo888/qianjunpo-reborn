// Package agones implements pre-prod Roma / 國戰 room lifecycle against Agones
// (mock in CI, sidecar REST + optional allocation HTTP on Dev clusters).
package agones

import "time"

// Phase is the Agones-aligned room lifecycle phase tracked locally.
type Phase string

const (
	PhaseUninitialized Phase = "uninitialized"
	PhaseAllocated     Phase = "allocated"
	PhaseReady         Phase = "ready"
	PhaseShuttingDown  Phase = "shutting_down"
	PhaseShutdown      Phase = "shutdown"
	PhaseError         Phase = "error"
)

// RoomIdentity identifies an allocated GameServer / Roma room instance.
type RoomIdentity struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Address   string `json:"address,omitempty"`
	Port      int    `json:"port,omitempty"`
	RoomID    string `json:"room_id,omitempty"`
}

// AllocateRequest selects a Fleet GameServer for a zone shard.
type AllocateRequest struct {
	ZoneID string
	Shard  uint32
}

// StatusSnapshot is returned by HTTP /v1/agones/status and tests.
type StatusSnapshot struct {
	Backend           string    `json:"backend"`
	AllocatorBackend  string    `json:"allocator_backend"`
	Phase             Phase     `json:"phase"`
	Identity          RoomIdentity `json:"identity"`
	LastError         string    `json:"last_error,omitempty"`
	LastTransitionUTC time.Time `json:"last_transition_utc,omitempty"`
	Ready             bool      `json:"ready"`
}

// OperationError carries observable failure metadata for logs and API responses.
type OperationError struct {
	Operation string
	Phase     Phase
	Err       error
	Timeout   bool
}

func (e *OperationError) Error() string {
	if e == nil || e.Err == nil {
		return "agones: unknown error"
	}
	if e.Timeout {
		return e.Operation + " timed out: " + e.Err.Error()
	}
	return e.Operation + " failed: " + e.Err.Error()
}

func (e *OperationError) Unwrap() error { return e.Err }
