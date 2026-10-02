package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rag"
)

func TestBattleContextChangesRAGHits(t *testing.T) {
	ctx := context.Background()
	store, err := rag.SeedStore(ctx, rag.NewHashBagEmbedder(64))
	if err != nil {
		t.Fatal(err)
	}
	order := "hold the center"
	with := RetrieveAndAugment(ctx, store, 5, "guard", order, BattleContext{
		UnitsSummary:   "cavalry flank ambush maneuver on the east",
		TerrainSummary: "hills high ground archer range",
	})
	if len(with.HitIDs) == 0 {
		t.Fatal("expected hits with battle context")
	}
	if with.HitIDs[0] != "tactic-flank" && with.HitIDs[0] != "tactic-hill" && with.HitIDs[0] != "wei-xiahou" {
		t.Fatalf("expected flank/hill/cavalry doc, got %v", with.HitIDs)
	}
	if !strings.Contains(with.AugmentedPrompt, "Battlefield:") {
		t.Fatalf("missing battlefield block: %q", with.AugmentedPrompt)
	}
}

func TestRAGInferWithBattleObservability(t *testing.T) {
	var gotPrompt string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Prompt string `json:"prompt"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotPrompt = body.Prompt
		_ = json.NewEncoder(w).Encode(RAGInferHTTPResponse("ok", "m", 1, 5, []string{"tactic-flank"}))
	}))
	defer srv.Close()

	client, err := DefaultRAGInferClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	off := false
	client.Infer.LogFallback = &off
	out := client.InferWithBattle(context.Background(), "guard", "press east", BattleContext{
		UnitsSummary: "cavalry waiting to flank",
	})
	if out.Text != "ok" || out.Source != SourceEdge {
		t.Fatalf("%+v", out)
	}
	if out.RAGK != 5 || len(out.RAGHitIDs) == 0 {
		t.Fatalf("rag observability: k=%d ids=%v", out.RAGK, out.RAGHitIDs)
	}
	if gotPrompt == "" || !strings.Contains(gotPrompt, "cavalry") {
		t.Fatalf("prompt=%q", gotPrompt)
	}
}
