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

type suggestRequest struct {
	BattleID      string           `json:"battle_id"`
	Persona       string           `json:"persona"`
	Order         string           `json:"order"`
	Prompt        string           `json:"prompt"`
	BattleContext ai.BattleContext `json:"battle_context"`
}

func (s *server) handleSuggest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req suggestRequest
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
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()
	inferFn := func(callCtx context.Context, prompt string) (string, string, int64, error) {
		out, err := s.backend.Infer(callCtx, backend.InferRequest{Prompt: prompt})
		if err != nil {
			return "", "", 0, err
		}
		return out.Text, out.Model, out.LatencyMs, nil
	}
	outcome, err := ai.RunStaggeredSuggest(ctx, store, rag.DefaultTopK, inferFn, ai.StaggeredSuggestInput{
		BattleID:      req.BattleID,
		Persona:       req.Persona,
		Order:         order,
		BattleContext: req.BattleContext,
	}, ai.GlobalSuggestionStore())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(inferErrorResponse{
			Error: err.Error(),
			Code:  "backend_infer_failed",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(outcome.InferHTTPResponse)
}

func (s *server) handleSuggestGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	battleID := r.URL.Query().Get("battle_id")
	if battleID == "" {
		http.Error(w, "battle_id required", http.StatusBadRequest)
		return
	}
	sug, ok := ai.GlobalSuggestionStore().GetSuggestion(battleID)
	if !ok {
		http.Error(w, "no suggestion for battle", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"battle_id":  battleID,
		"suggestion": sug,
	})
}
