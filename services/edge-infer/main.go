// Package main is the edge inference service (Qwen2.5-3B mock or Ollama backend).
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/ai"
	"github.com/iiooiioo888/qianjunpo-reborn/services/edge-infer/backend"
)

type inferRequest struct {
	Prompt string `json:"prompt"`
}

type inferErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

type server struct {
	backend backend.Backend
	mock    *backend.MockBackend
}

func main() {
	cfg := loadConfig()
	be, mock := newBackend(cfg)
	srv := &server{backend: be, mock: mock}
	addr := cfg.Addr
	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.handleHealth)
	mux.HandleFunc("/v1/infer", srv.handleInfer)
	mux.HandleFunc("/v1/load", srv.handleLoad)
	log.Printf("edge-infer listening on %s backend=%s model=%s", addr, be.Name(), be.Model())
	log.Fatal(http.ListenAndServe(addr, mux))
}

type healthResponse struct {
	backend.Health
	NPCFallbackByReason map[string]uint64 `json:"npc_fallback_by_reason"`
}

func (s *server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	h := s.backend.Health(ctx)
	payload := healthResponse{
		Health:              h,
		NPCFallbackByReason: ai.NPCFallbackCounts(),
	}
	w.Header().Set("Content-Type", "application/json")
	if h.Ready || h.Backend == "mock" {
		w.WriteHeader(http.StatusOK)
	} else {
		// Soft fail: process stays up; callers may use NPC fallback.
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *server) handleLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.mock != nil {
		s.mock.Heavy.Store(true)
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *server) handleInfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req inferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()
	out, err := s.backend.Infer(ctx, backend.InferRequest{Prompt: req.Prompt})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(inferErrorResponse{
			Error: err.Error(),
			Code:  "backend_infer_failed",
		})
		return
	}
	resp := ai.EdgeInferHTTPResponse(out.Text, out.Model, out.LatencyMs)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
