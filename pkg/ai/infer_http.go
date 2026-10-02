package ai

// InferHTTPResponse is the JSON body for a successful POST /v1/infer (edge-infer).
type InferHTTPResponse struct {
	Text           string           `json:"text"`
	Model          string           `json:"model"`
	LatencyMs      int64            `json:"latency_ms"`
	Source         string           `json:"source"`
	RAGK           int              `json:"rag_k,omitempty"`
	RAGHitIDs      []string         `json:"rag_hit_ids,omitempty"`
	FallbackReason string           `json:"fallback_reason,omitempty"`
	FallbackDetail string           `json:"fallback_detail,omitempty"`
	Suggestion     *InferSuggestion `json:"suggestion,omitempty"`
}

// InferPostBody is the JSON body pkg/ai sends to POST /v1/infer.
type InferPostBody struct {
	Prompt    string   `json:"prompt"`
	RAGK      int      `json:"rag_k,omitempty"`
	RAGHitIDs []string `json:"rag_hit_ids,omitempty"`
}

// EdgeInferHTTPResponse builds the success contract for edge model output.
func EdgeInferHTTPResponse(text, model string, latencyMs int64) InferHTTPResponse {
	return EdgeInferHTTPResponseWithRAG(text, model, latencyMs, 0, nil)
}

// EdgeInferHTTPResponseWithRAG echoes RAG participation on the wire when the client supplied Top-K metadata.
func EdgeInferHTTPResponseWithRAG(text, model string, latencyMs int64, ragK int, ragHitIDs []string) InferHTTPResponse {
	return InferHTTPResponse{
		Text:      text,
		Model:     model,
		LatencyMs: latencyMs,
		Source:    SourceEdge,
		RAGK:      ragK,
		RAGHitIDs: ragHitIDs,
	}
}

// InferResultToHTTP maps in-process InferResult into the shared POST /v1/infer success JSON shape.
// Gateways and Live HUD should use the same snake_case fields as InferHTTPResponse.
func InferResultToHTTP(r InferResult, model string, latencyMs int64) InferHTTPResponse {
	source := r.Source
	if source == "" {
		source = SourceEdge
	}
	resp := InferHTTPResponse{
		Text:      r.Text,
		Model:     model,
		LatencyMs: latencyMs,
		Source:    source,
		RAGK:      r.RAGK,
		RAGHitIDs: append([]string(nil), r.RAGHitIDs...),
	}
	if r.Suggestion != nil {
		cp := *r.Suggestion
		resp.Suggestion = &cp
	}
	if source == SourceNPC && r.FallbackReason != FallbackReasonNone {
		resp.FallbackReason = string(r.FallbackReason)
		resp.FallbackDetail = r.FallbackDetail
	}
	return resp
}

// InferResultFromHTTP maps a decoded success response into InferResult.
// Empty text yields a zero result (caller should treat as bad response).
func InferResultFromHTTP(r InferHTTPResponse) InferResult {
	if r.Text == "" {
		return InferResult{}
	}
	source := r.Source
	if source == "" {
		source = SourceEdge
	}
	reason := FallbackReason(r.FallbackReason)
	if source == SourceEdge {
		reason = FallbackReasonNone
	}
	out := InferResult{
		Text:           r.Text,
		Source:         source,
		RAGK:           r.RAGK,
		RAGHitIDs:      append([]string(nil), r.RAGHitIDs...),
		FallbackReason: reason,
		FallbackDetail: r.FallbackDetail,
	}
	if r.Suggestion != nil {
		cp := *r.Suggestion
		out.Suggestion = &cp
	}
	return out
}

// RAGInferHTTPResponse builds the success contract for POST /v1/rag-infer.
func RAGInferHTTPResponse(text, model string, latencyMs int64, ragK int, hitIDs []string) InferHTTPResponse {
	return EdgeInferHTTPResponseWithRAG(text, model, latencyMs, ragK, hitIDs)
}
