package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRAGInferNoInferClientIncrementsCounter(t *testing.T) {
	ResetNPCFallbackCounts()
	client, err := DefaultRAGInferClient("http://127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	client.Infer = nil
	reason := string(FallbackReasonNoInferClient)
	before := NPCFallbackCounts()[reason]
	out := client.InferWithContext(context.Background(), "guard", "hold the line")
	if out.Source != SourceNPC || out.FallbackReason != FallbackReasonNoInferClient {
		t.Fatalf("%+v", out)
	}
	after := NPCFallbackCounts()[reason]
	if after <= before {
		t.Fatalf("counter for %q did not increase: before=%d after=%d", reason, before, after)
	}
}

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
	off := false
	client.Infer.LogFallback = &off

	start := time.Now()
	out := client.InferWithContext(context.Background(), "guard", "flank the supply line")
	elapsed := time.Since(start)
	if out.Text != NPCFallback("guard") {
		t.Fatalf("expected NPC fallback, got %q", out.Text)
	}
	if out.Source != SourceNPC || out.FallbackReason != FallbackReasonTimeout {
		t.Fatalf("%+v", out)
	}
	if elapsed > 600*time.Millisecond {
		t.Fatalf("RAG+infer fallback too slow: %v", elapsed)
	}
}
