// Package rag defines a retrieval layer for strategic AI context (Top-K).
//
// Latency target: in-memory backend should answer Top-5 queries in <50ms P99 on dev hardware;
// Milvus/Qdrant clusters are out of scope for this phase (see docs/ai-edge-rag.md).
package rag

import (
	"context"
	"sort"
	"strings"
	"sync"
)

// Document is one retrievable chunk.
type Document struct {
	ID      string
	Content string
	Faction string
	General string
	Vector  []float32
	RawText string
}

// Store retrieves similar documents (vector search).
type Store interface {
	Upsert(ctx context.Context, docs ...Document) error
	TopK(ctx context.Context, query []float32, k int) ([]Document, error)
	TopKText(ctx context.Context, query string, k int) ([]Document, error)
}

// InMemoryStore is a cosine-similarity backend for tests and offline dev.
type InMemoryStore struct {
	mu       sync.RWMutex
	docs     []Document
	embedder Embedder
}

// Upsert appends documents (ids may repeat in this simple store).
func (s *InMemoryStore) Upsert(_ context.Context, docs ...Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.docs = append(s.docs, docs...)
	return nil
}

func (s *InMemoryStore) TopK(_ context.Context, query []float32, k int) ([]Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return topKCosine(s.docs, query, k), nil
}

func (s *InMemoryStore) TopKText(ctx context.Context, query string, k int) ([]Document, error) {
	if s.embedder == nil {
		s.embedder = NewHashBagEmbedder(64)
	}
	vec := s.embedder.Embed(query)
	return s.TopK(ctx, vec, k)
}

func topKCosine(docs []Document, query []float32, k int) []Document {
	if k <= 0 {
		k = DefaultTopK
	}
	type scored struct {
		doc Document
		sim float64
	}
	scores := make([]scored, 0, len(docs))
	for _, d := range docs {
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
	return out
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
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 8; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

// FormatContext joins Top-K hits into a prompt prefix (does not affect sim math).
func FormatContext(docs []Document) string {
	if len(docs) == 0 {
		return ""
	}
	var b strings.Builder
	for i, d := range docs {
		if i > 0 {
			b.WriteString("\n")
		}
		if d.General != "" {
			b.WriteString("- [")
			b.WriteString(d.Faction)
			b.WriteString("·")
			b.WriteString(d.General)
			b.WriteString("] ")
		}
		b.WriteString(d.Content)
	}
	return b.String()
}
