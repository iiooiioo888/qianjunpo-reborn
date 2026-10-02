package ai

// InferHTTPResponse is the JSON body for a successful POST /v1/infer (edge-infer).
type InferHTTPResponse struct {
	Text           string   `json:"text"`
	Model          string   `json:"model"`
	LatencyMs      int64    `json:"latency_ms"`
	Source         string   `json:"source"`
	RAGK           int      `json:"rag_k,omitempty"`
	RAGHitIDs      []string `json:"rag_hit_ids,omitempty"`
	FallbackReason string   `json:"fallback_reason,omitempty"`
	FallbackDetail string   `json:"fallback_detail,omitempty"`
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
	return InferResult{
		Text:           r.Text,
		Source:         source,
		RAGK:           r.RAGK,
		RAGHitIDs:      append([]string(nil), r.RAGHitIDs...),
		FallbackReason: reason,
		FallbackDetail: r.FallbackDetail,
	}
}
