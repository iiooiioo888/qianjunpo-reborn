package rag

import (
	"context"
	"testing"
)

func TestInMemoryTop5(t *testing.T) {
	ctx := context.Background()
	s := &InMemoryStore{}
	_ = s.Upsert(ctx,
		Document{ID: "a", Content: "spear wall", Vector: []float32{1, 0, 0}},
		Document{ID: "b", Content: "cavalry flank", Vector: []float32{0.9, 0.1, 0}},
		Document{ID: "c", Content: "archer hill", Vector: []float32{0, 1, 0}},
	)
	hits, err := s.TopK(ctx, []float32{1, 0, 0}, 5)
	if err != nil || len(hits) == 0 || hits[0].ID != "a" {
		t.Fatalf("hits=%v err=%v", hits, err)
	}
}
