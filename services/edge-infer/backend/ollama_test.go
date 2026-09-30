package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOllamaInferOpenAICompat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"choices": []map[string]interface{}{
					{"message": map[string]string{"content": "retreat to hill"}},
				},
			})
			return
		}
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"models":[]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	be := &OllamaBackend{BaseURL: srv.URL, ModelName: "qwen2.5:3b"}
	out, err := be.Infer(context.Background(), InferRequest{Prompt: "hold line"})
	if err != nil || out.Text != "retreat to hill" {
		t.Fatalf("out=%+v err=%v", out, err)
	}
	h := be.Health(context.Background())
	if !h.Ready || h.Backend != "ollama" {
		t.Fatalf("%+v", h)
	}
}

func TestOllamaHealthDegraded(t *testing.T) {
	be := &OllamaBackend{BaseURL: "http://127.0.0.1:1", ModelName: "qwen2.5:3b", HTTPClient: &http.Client{Timeout: 50 * time.Millisecond}}
	h := be.Health(context.Background())
	if h.Ready || h.Status != "degraded" {
		t.Fatalf("%+v", h)
	}
}

func TestOllamaHealthProbeTimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			<-r.Context().Done()
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	be := &OllamaBackend{
		BaseURL:            srv.URL,
		ModelName:          "qwen2.5:3b",
		HealthProbeTimeout: 80 * time.Millisecond,
	}
	start := time.Now()
	h := be.Health(context.Background())
	elapsed := time.Since(start)
	if h.Ready || h.Status != "degraded" {
		t.Fatalf("%+v", h)
	}
	if elapsed > 400*time.Millisecond {
		t.Fatalf("health probe should fail fast, took %v", elapsed)
	}
}
