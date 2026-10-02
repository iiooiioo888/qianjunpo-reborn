package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rag"
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

func TestRAGInferTopKEntersPromptAndResult(t *testing.T) {
	var post InferPostBody
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			t.Errorf("decode: %v", err)
		}
		_ = json.NewEncoder(w).Encode(EdgeInferHTTPResponseWithRAG("ok", "qwen2.5-3b-mock", 1, post.RAGK, post.RAGHitIDs))
	}))
	defer srv.Close()

	client, err := DefaultRAGInferClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	off := false
	client.Infer.LogFallback = &off
	client.LogRAG = &off
	client.TopK = 3

	query := "flank cavalry ambush maneuver"
	out := client.InferWithContext(context.Background(), "cavalry", query)
	if out.Text != "ok" || out.Source != SourceEdge {
		t.Fatalf("%+v", out)
	}
	if out.RAGK != 3 || len(out.RAGHitIDs) == 0 {
		t.Fatalf("rag observability: k=%d ids=%v", out.RAGK, out.RAGHitIDs)
	}
	if !strings.Contains(post.Prompt, "Knowledge:") || !strings.Contains(post.Prompt, query) {
		t.Fatalf("prompt missing RAG block: %q", post.Prompt)
	}
	if !strings.Contains(post.Prompt, "側翼包抄") {
		t.Fatalf("expected flank tactic chunk in prompt, got %q", post.Prompt)
	}
	foundFlank := false
	for _, id := range out.RAGHitIDs {
		if id == "tactic-flank" {
			foundFlank = true
			break
		}
	}
	if !foundFlank {
		t.Fatalf("expected tactic-flank in hits, got %v", out.RAGHitIDs)
	}
	if post.RAGK != 3 || len(post.RAGHitIDs) != len(out.RAGHitIDs) {
		t.Fatalf("post body rag meta: %+v", post)
	}
}

func TestRAGInferLogsAugmentation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(EdgeInferHTTPResponse("ok", "m", 1))
	}))
	defer srv.Close()

	client, err := DefaultRAGInferClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	off := false
	client.Infer.LogFallback = &off

	logs := withSlogCapture(t, func() {
		out := client.InferWithContext(context.Background(), "scout", "flank cavalry ambush")
		if out.Source != SourceEdge {
			t.Fatalf("%+v", out)
		}
	})
	if !strings.Contains(logs, "ai infer rag augmented") {
		t.Fatalf("missing rag log: %q", logs)
	}
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(logs)), &entry); err != nil {
		t.Fatal(err)
	}
	if entry["rag_k"] == nil {
		t.Fatalf("rag_k missing: %v", entry)
	}
	ids, ok := entry["rag_hit_ids"].([]interface{})
	if !ok || len(ids) == 0 {
		t.Fatalf("rag_hit_ids=%v", entry["rag_hit_ids"])
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
	client.LogRAG = &off

	start := time.Now()
	out := client.InferWithContext(context.Background(), "guard", "flank the supply line")
	elapsed := time.Since(start)
	if out.Text != NPCFallback("guard") {
		t.Fatalf("expected NPC fallback, got %q", out.Text)
	}
	if out.Source != SourceNPC || out.FallbackReason != FallbackReasonTimeout {
		t.Fatalf("%+v", out)
	}
	if out.RAGK != rag.DefaultTopK || len(out.RAGHitIDs) == 0 {
		t.Fatalf("expected RAG meta on fallback path: k=%d ids=%v", out.RAGK, out.RAGHitIDs)
	}
	if elapsed > 600*time.Millisecond {
		t.Fatalf("RAG+infer fallback too slow: %v", elapsed)
	}
}
