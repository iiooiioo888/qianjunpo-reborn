package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/ai"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rag"
	"github.com/iiooiioo888/qianjunpo-reborn/services/edge-infer/backend"
)

type ragInferRequest struct {
	Persona       string            `json:"persona"`
	Order         string            `json:"order"`
	Prompt        string            `json:"prompt"` // alias for order (curl convenience)
	BattleContext ai.BattleContext  `json:"battle_context"`
}

func (s *server) ensureRAGStore() (rag.Store, error) {
	if s.ragStore != nil {
		return s.ragStore, nil
	}
	store, err := rag.SeedStore(context.Background(), rag.NewHashBagEmbedder(64))
	if err != nil {
		return nil, err
	}
	s.ragStore = store
	return store, nil
}

func (s *server) handleRAGInfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ragInferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	order := req.Order
	if order == "" {
		order = req.Prompt
	}
	if order == "" {
		http.Error(w, "order or prompt required", http.StatusBadRequest)
		return
	}
	store, err := s.ensureRAGStore()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	topK := rag.DefaultTopK
	retrieved := ai.RetrieveAndAugment(r.Context(), store, topK, req.Persona, order, req.BattleContext)

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()
	out, err := s.backend.Infer(ctx, backend.InferRequest{Prompt: retrieved.AugmentedPrompt})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(inferErrorResponse{
			Error: err.Error(),
			Code:  "backend_infer_failed",
		})
		return
	}
	resp := ai.RAGInferHTTPResponse(out.Text, out.Model, out.LatencyMs, retrieved.K, retrieved.HitIDs)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
