package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestInferClientMockHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/infer" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"text": "advance: test", "model": "qwen2.5-3b-mock", "latency_ms": 1,
		})
	}))
	defer srv.Close()

	c := &InferClient{BaseURL: srv.URL}
	out := c.Infer(context.Background(), "guard", "defend gate")
	if out.Text != "advance: test" {
		t.Fatalf("%q", out.Text)
	}
}

func TestInferClientNPCFallback(t *testing.T) {
	c := &InferClient{BaseURL: "http://127.0.0.1:1"}
	out := c.Infer(context.Background(), "guard", "defend")
	if out.Text != NPCFallback("guard") {
		t.Fatalf("%q", out.Text)
	}
}

func TestInferClientSlowInferFallsBackWithinDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Slower than client HTTPTimeout; short enough not to stall test teardown.
		time.Sleep(300 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "late"})
	}))
	defer srv.Close()

	c := &InferClient{
		BaseURL:        srv.URL,
		RequestTimeout: 200 * time.Millisecond,
		HTTPTimeout:    150 * time.Millisecond,
	}
	start := time.Now()
	out := c.Infer(context.Background(), "guard", "hold")
	elapsed := time.Since(start)
	if out.Text != NPCFallback("guard") {
		t.Fatalf("expected NPC fallback, got %q", out.Text)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("fallback took too long: %v", elapsed)
	}
}

func TestRAGInferWiring(t *testing.T) {
	var gotPrompt string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Prompt string `json:"prompt"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotPrompt = body.Prompt
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "ok"})
	}))
	defer srv.Close()

	client, err := DefaultRAGInferClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	out := client.InferWithContext(context.Background(), "guard", "flank east with cavalry")
	if out.Text != "ok" {
		t.Fatalf("%q", out.Text)
	}
	if gotPrompt == "" || !strings.Contains(gotPrompt, "Knowledge:") || !strings.Contains(gotPrompt, "flank") || !strings.Contains(gotPrompt, "Order:") {
		t.Fatalf("prompt=%q", gotPrompt)
	}
}
