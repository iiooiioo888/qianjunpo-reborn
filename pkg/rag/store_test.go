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

func TestSeedCorpusTopKMeaningful(t *testing.T) {
	ctx := context.Background()
	store, err := SeedStore(ctx, NewHashBagEmbedder(64))
	if err != nil {
		t.Fatal(err)
	}
	hits, err := store.TopKText(ctx, "cavalry flank ambush maneuver", 5)
	if err != nil || len(hits) == 0 {
		t.Fatalf("hits=%v err=%v", hits, err)
	}
	if hits[0].ID != "tactic-flank" && hits[0].ID != "wei-xiahou" {
		t.Fatalf("expected flank-related doc first, got %s (%s)", hits[0].ID, hits[0].Content)
	}
}

func TestHashBagEmbedderDeterministic(t *testing.T) {
	e := NewHashBagEmbedder(32)
	a := e.Embed("zhang fei bridge")
	b := e.Embed("zhang fei bridge")
	if len(a) != 32 || len(b) != 32 {
		t.Fatal(len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("non-deterministic embed")
		}
	}
}
