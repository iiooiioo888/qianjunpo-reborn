package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
