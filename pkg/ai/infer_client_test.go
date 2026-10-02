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

func inferClientNoLog(base string, extra func(*InferClient)) *InferClient {
	off := false
	c := &InferClient{BaseURL: base, LogFallback: &off}
	if extra != nil {
		extra(c)
	}
	return c
}

func TestInferClientMockHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/infer" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(EdgeInferHTTPResponse("advance: test", "qwen2.5-3b-mock", 1))
	}))
	defer srv.Close()

	c := inferClientNoLog(srv.URL, nil)
	out := c.Infer(context.Background(), "guard", "defend gate")
	if out.Text != "advance: test" || out.Source != SourceEdge || out.FallbackReason != FallbackReasonNone {
		t.Fatalf("%+v", out)
	}
}

func TestInferClientDecodesHTTPObservabilityFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(EdgeInferHTTPResponse("line holds", "qwen2.5-3b-mock", 2))
	}))
	defer srv.Close()

	c := inferClientNoLog(srv.URL, nil)
	out := c.Infer(context.Background(), "guard", "hold")
	if out.Source != SourceEdge || out.FallbackReason != FallbackReasonNone || out.Text != "line holds" {
		t.Fatalf("%+v", out)
	}
}

func TestInferClientNPCFallback(t *testing.T) {
	c := inferClientNoLog("http://127.0.0.1:1", nil)
	out := c.Infer(context.Background(), "guard", "defend")
	if out.Text != NPCFallback("guard") || out.Source != SourceNPC {
		t.Fatalf("%+v", out)
	}
	if out.FallbackReason != FallbackReasonHTTPTransport && out.FallbackReason != FallbackReasonTimeout {
		t.Fatalf("unexpected reason %q", out.FallbackReason)
	}
}

func TestInferClientFallbackReasonHTTPStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "model down",
			"code":  "backend_infer_failed",
		})
	}))
	defer srv.Close()

	c := inferClientNoLog(srv.URL, nil)
	out := c.Infer(context.Background(), "scout", "probe north")
	if out.Source != SourceNPC || out.FallbackReason != FallbackReasonHTTPStatus {
		t.Fatalf("%+v", out)
	}
	if out.Text != NPCFallback("scout") {
		t.Fatalf("text %q", out.Text)
	}
	if !strings.Contains(out.FallbackDetail, "502") || !strings.Contains(out.FallbackDetail, "backend_infer_failed") {
		t.Fatalf("detail %q", out.FallbackDetail)
	}
	raw, err := json.Marshal(InferResultToHTTP(out, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]interface{}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	if wire["source"] != SourceNPC || wire["fallback_reason"] != string(FallbackReasonHTTPStatus) {
		t.Fatalf("HUD wire JSON: %v", wire)
	}
}

func TestInferClientFallbackReasonBadResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"text":""}`))
	}))
	defer srv.Close()

	c := inferClientNoLog(srv.URL, nil)
	out := c.Infer(context.Background(), "strategist", "hold")
	if out.FallbackReason != FallbackReasonBadResponse || out.Source != SourceNPC {
		t.Fatalf("%+v", out)
	}
}

func TestInferClientSlowInferFallsBackWithinDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Slower than client HTTPTimeout; short enough not to stall test teardown.
		time.Sleep(300 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "late"})
	}))
	defer srv.Close()

	c := inferClientNoLog(srv.URL, func(c *InferClient) {
		c.RequestTimeout = 200 * time.Millisecond
		c.HTTPTimeout = 150 * time.Millisecond
	})
	start := time.Now()
	out := c.Infer(context.Background(), "guard", "hold")
	elapsed := time.Since(start)
	if out.Text != NPCFallback("guard") {
		t.Fatalf("expected NPC fallback, got %q", out.Text)
	}
	if out.FallbackReason != FallbackReasonTimeout {
		t.Fatalf("reason=%q", out.FallbackReason)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("fallback took too long: %v", elapsed)
	}
}

func TestRAGInferWiring(t *testing.T) {
	var gotPrompt string
	var post InferPostBody
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&post)
		gotPrompt = post.Prompt
		_ = json.NewEncoder(w).Encode(EdgeInferHTTPResponseWithRAG("ok", "m", 1, post.RAGK, post.RAGHitIDs))
	}))
	defer srv.Close()

	client, err := DefaultRAGInferClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	off := false
	client.Infer.LogFallback = &off
	client.LogRAG = &off
	out := client.InferWithContext(context.Background(), "guard", "flank east with cavalry")
	if out.Text != "ok" || out.Source != SourceEdge {
		t.Fatalf("%+v", out)
	}
	if gotPrompt == "" || !strings.Contains(gotPrompt, "Knowledge:") || !strings.Contains(gotPrompt, "flank") || !strings.Contains(gotPrompt, "Order:") {
		t.Fatalf("prompt=%q", gotPrompt)
	}
	if out.RAGK != rag.DefaultTopK || len(out.RAGHitIDs) == 0 {
		t.Fatalf("rag meta: k=%d ids=%v", out.RAGK, out.RAGHitIDs)
	}
}

func TestRAGInferNoInferClient(t *testing.T) {
	logs := withSlogCapture(t, func() {
		store, err := DefaultRAGInferClient("http://example.com")
		if err != nil {
			t.Fatal(err)
		}
		store.Infer = nil
		out := store.InferWithContext(context.Background(), "merchant", "escort wagons")
		if out.Source != SourceNPC || out.FallbackReason != FallbackReasonNoInferClient {
			t.Fatalf("%+v", out)
		}
		if out.Text != NPCFallback("merchant") {
			t.Fatalf("%q", out.Text)
		}
	})
	if !strings.Contains(logs, "fallback_reason") {
		t.Fatalf("expected structured fallback log, got %q", logs)
	}
}
