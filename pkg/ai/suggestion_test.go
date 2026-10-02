package ai

import (
	"context"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rag"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

func TestDeriveSuggestionMoveFromAdvance(t *testing.T) {
	sug := DeriveSuggestion(InferPathEdge, "advance: hold the gate", nil)
	if sug.Kind != SuggestionKindMove || sug.ToX != 5 || sug.ToY != 8 {
		t.Fatalf("%+v", sug)
	}
	if sug.UnitID != tactical.UnitIDPlayer0 {
		t.Fatalf("unit=%d", sug.UnitID)
	}
}

func TestDeriveSuggestionSkillFromRAGSpear(t *testing.T) {
	sug := DeriveSuggestion(InferPathRAG, "advance: flank", []string{"tactic-spear", "wei-cao"})
	if sug.Kind != SuggestionKindSkill || sug.SkillID != 1 {
		t.Fatalf("%+v", sug)
	}
	if sug.InferPath != InferPathRAG {
		t.Fatalf("path=%s", sug.InferPath)
	}
}

func TestStaggerAlternatesPaths(t *testing.T) {
	ResetStaggerForTests()
	if NextStaggerPath() != InferPathRAG {
		t.Fatal("first leg should be rag")
	}
	if NextStaggerPath() != InferPathEdge {
		t.Fatal("second leg should be edge")
	}
}

func TestSuggestionStoreWriteBack(t *testing.T) {
	store := &SuggestionStore{byID: make(map[string]InferSuggestion)}
	sug := moveSuggestion(InferPathEdge, 5, 8)
	store.WriteSuggestion("default/0", sug)
	got, ok := store.GetSuggestion("default/0")
	if !ok || got.ToX != 5 {
		t.Fatalf("ok=%v got=%+v", ok, got)
	}
}

func TestRunStaggeredSuggestWriteBack(t *testing.T) {
	ResetStaggerForTests()
	ForceStaggerPath(InferPathEdge)
	ctx := context.Background()
	embed := rag.NewHashBagEmbedder(64)
	ragStore, err := rag.SeedStore(ctx, embed)
	if err != nil {
		t.Fatal(err)
	}
	sb := &SuggestionStore{byID: make(map[string]InferSuggestion)}
	inferFn := func(_ context.Context, prompt string) (string, string, int64, error) {
		return "advance: " + prompt, "mock", 1, nil
	}
	out, err := RunStaggeredSuggest(ctx, ragStore, rag.DefaultTopK, inferFn, StaggeredSuggestInput{
		BattleID: "b1",
		Persona:  "guard",
		Order:    "hold east",
	}, sb)
	if err != nil {
		t.Fatal(err)
	}
	if out.Suggestion.Kind != SuggestionKindMove {
		t.Fatalf("%+v", out.Suggestion)
	}
	if out.InferHTTPResponse.Suggestion == nil || out.InferHTTPResponse.Suggestion.Kind != SuggestionKindMove {
		t.Fatalf("http sug=%+v", out.InferHTTPResponse.Suggestion)
	}
	got, ok := sb.GetSuggestion("b1")
	if !ok || got.Kind != SuggestionKindMove {
		t.Fatalf("write-back missing: %+v", got)
	}
}
