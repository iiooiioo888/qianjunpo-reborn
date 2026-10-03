package ai

import (
	"context"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rag"
)

// EdgeInferFunc runs a single backend infer for an augmented or plain prompt.
type EdgeInferFunc func(ctx context.Context, prompt string) (text, model string, latencyMs int64, err error)

// StaggeredSuggestInput is the HTTP/curl body for suggest write-back.
type StaggeredSuggestInput struct {
	BattleID      string
	Persona       string
	Order         string
	BattleContext BattleContext
}

// StaggeredSuggestOutcome merges infer observability with one tactical suggestion.
type StaggeredSuggestOutcome struct {
	InferHTTPResponse InferHTTPResponse
	Suggestion        InferSuggestion
}

// RunStaggeredSuggest picks edge vs RAG path, infers, derives suggestion, optional write-back.
func RunStaggeredSuggest(
	ctx context.Context,
	store rag.Store,
	topK int,
	infer EdgeInferFunc,
	in StaggeredSuggestInput,
	writeBack *SuggestionStore,
) (StaggeredSuggestOutcome, error) {
	path := NextStaggerPath()
	prompt := in.Order
	var ragK int
	var hitIDs []string
	if path == InferPathRAG && store != nil {
		retrieved := RetrieveAndAugment(ctx, store, topK, in.Persona, in.Order, in.BattleContext)
		prompt = retrieved.AugmentedPrompt
		ragK = retrieved.K
		hitIDs = retrieved.HitIDs
	}
	text, model, latencyMs, err := infer(ctx, prompt)
	var resp InferHTTPResponse
	if err != nil {
		npc := npcFallbackResult(in.Persona, FallbackReasonHTTPStatus, err.Error(), true)
		text = npc.Text
		sug := DeriveSuggestion(path, text, hitIDs)
		resp = InferResultToHTTP(npc.WithRAGObservability(ragK, hitIDs), "", 0)
		resp.Suggestion = &sug
		AttachTacticalCommand(&resp, in.BattleID)
		if writeBack != nil && in.BattleID != "" {
			writeBack.WriteSuggestion(in.BattleID, sug)
		}
		return StaggeredSuggestOutcome{InferHTTPResponse: resp, Suggestion: sug}, nil
	}
	sug := DeriveSuggestion(path, text, hitIDs)
	if writeBack != nil && in.BattleID != "" {
		writeBack.WriteSuggestion(in.BattleID, sug)
	}
	resp = EdgeInferHTTPResponseWithRAG(text, model, latencyMs, ragK, hitIDs)
	resp.Suggestion = &sug
	AttachTacticalCommand(&resp, in.BattleID)
	return StaggeredSuggestOutcome{InferHTTPResponse: resp, Suggestion: sug}, nil
}
