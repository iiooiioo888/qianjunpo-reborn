package rag

import (
	"hash/fnv"
	"strings"
)

// Embedder maps text to a dense vector for similarity search.
// Remote providers (OpenAI, etc.) can implement the same interface later.
type Embedder interface {
	Embed(text string) []float32
	Dim() int
}

// HashBagEmbedder is a deterministic offline embedder (token hashing + L2 norm).
type HashBagEmbedder struct {
	dim int
}

// NewHashBagEmbedder returns an embedder with fixed dimension (default 64).
func NewHashBagEmbedder(dim int) *HashBagEmbedder {
	if dim <= 0 {
		dim = 64
	}
	return &HashBagEmbedder{dim: dim}
}

func (e *HashBagEmbedder) Dim() int { return e.dim }

func (e *HashBagEmbedder) Embed(text string) []float32 {
	vec := make([]float32, e.dim)
	for _, tok := range tokenize(text) {
		h := fnv.New32a()
		_, _ = h.Write([]byte(tok))
		idx := int(h.Sum32() % uint32(e.dim))
		vec[idx] += 1
	}
	norm := float32(0)
	for _, v := range vec {
		norm += v * v
	}
	if norm == 0 {
		return vec
	}
	inv := 1 / sqrt(float64(norm))
	for i := range vec {
		vec[i] = float32(float64(vec[i]) * inv)
	}
	return vec
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	var out []string
	for _, part := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '，' || r == '。' || r == '、' || r == ':' || r == ';'
	}) {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
