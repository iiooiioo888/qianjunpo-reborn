package ai

import "sync/atomic"

var staggerSeq atomic.Uint64

// NextStaggerPath alternates edge-only and RAG-augmented infer legs (process-wide).
func NextStaggerPath() string {
	if staggerSeq.Add(1)%2 == 0 {
		return InferPathEdge
	}
	return InferPathRAG
}

// ResetStaggerForTests clears the stagger counter (next leg is RAG).
func ResetStaggerForTests() {
	staggerSeq.Store(0)
}

// ForceStaggerPath sets the next NextStaggerPath result (tests only).
func ForceStaggerPath(path string) {
	if path == InferPathEdge {
		staggerSeq.Store(1)
		return
	}
	staggerSeq.Store(0)
}
