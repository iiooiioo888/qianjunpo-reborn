// Package rag defines a retrieval stub for strategic AI context (Top-K).
//
// Latency target: in-memory backend should answer Top-5 queries in <50ms P99 on dev hardware;
// production Milvus/Qdrant clusters are out of scope for this phase.
package rag

import (
	"context"
	"sort"
	"sync"
)

// Document is one retrievable chunk.
type Document struct {
	ID      string
	Content string
	Vector  []float32
}

// Store retrieves similar documents (vector search).
type Store interface {
	Upsert(ctx context.Context, docs ...Document) error
	TopK(ctx context.Context, query []float32, k int) ([]Document, error)
}

// InMemoryStore is a fake cosine-similarity backend for tests.
type InMemoryStore struct {
	mu   sync.RWMutex
	docs []Document
}

func (s *InMemoryStore) Upsert(_ context.Context, docs ...Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.docs = append(s.docs, docs...)
	return nil
}

func (s *InMemoryStore) TopK(_ context.Context, query []float32, k int) ([]Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if k <= 0 {
		k = 5
	}
	type scored struct {
		doc Document
		sim float64
	}
	scores := make([]scored, 0, len(s.docs))
	for _, d := range s.docs {
		scores = append(scores, scored{doc: d, sim: cosine(query, d.Vector)})
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].sim > scores[j].sim })
	if len(scores) > k {
		scores = scores[:k]
	}
	out := make([]Document, len(scores))
	for i, sc := range scores {
		out[i] = sc.doc
	}
	return out, nil
}

func cosine(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dot, na, nb float64
	for i := 0; i < n; i++ {
		dot += float64(a[i] * b[i])
		na += float64(a[i] * a[i])
		nb += float64(b[i] * b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (sqrt(na) * sqrt(nb))
}

func sqrt(x float64) float64 {
	// Newton iteration without importing math for tiny helper (tests only precision).
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 8; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}
