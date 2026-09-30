// Package main is a lightweight edge inference skeleton (Qwen2.5-3B mock).
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

type inferRequest struct {
	Prompt string `json:"prompt"`
}

type inferResponse struct {
	Text     string `json:"text"`
	Model    string `json:"model"`
	LatencyMs int64 `json:"latency_ms"`
}

var loadHook atomic.Bool

func main() {
	addr := env("EDGE_INFER_ADDR", ":8088")
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/v1/infer", handleInfer)
	mux.HandleFunc("/v1/load", handleLoad)
	log.Printf("edge-infer listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","model":"qwen2.5-3b-mock"}`))
}

func handleLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	loadHook.Store(true)
	w.WriteHeader(http.StatusAccepted)
}

func handleInfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	start := time.Now()
	var req inferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	text := mockQwen(req.Prompt, loadHook.Load())
	resp := inferResponse{
		Text:      text,
		Model:     "qwen2.5-3b-mock",
		LatencyMs: time.Since(start).Milliseconds(),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func mockQwen(prompt string, heavy bool) string {
	if prompt == "" {
		return "hold position"
	}
	if heavy {
		return "flank east with cavalry: " + prompt
	}
	return "advance: " + prompt
}
