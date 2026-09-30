package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRAGInferSlowEdgeFallsBackWithinDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, err := DefaultRAGInferClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	client.Infer.RequestTimeout = 250 * time.Millisecond
	client.Infer.HTTPTimeout = 200 * time.Millisecond

	start := time.Now()
	out := client.InferWithContext(context.Background(), "guard", "flank the supply line")
	elapsed := time.Since(start)
	if out.Text != NPCFallback("guard") {
		t.Fatalf("expected NPC fallback, got %q", out.Text)
	}
	if elapsed > 600*time.Millisecond {
		t.Fatalf("RAG+infer fallback too slow: %v", elapsed)
	}
}
