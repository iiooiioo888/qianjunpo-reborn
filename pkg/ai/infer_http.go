package ai

// InferHTTPResponse is the JSON body for a successful POST /v1/infer (edge-infer).
type InferHTTPResponse struct {
	Text           string `json:"text"`
	Model          string `json:"model"`
	LatencyMs      int64  `json:"latency_ms"`
	Source         string `json:"source"`
	FallbackReason string `json:"fallback_reason,omitempty"`
	FallbackDetail string `json:"fallback_detail,omitempty"`
}

// EdgeInferHTTPResponse builds the success contract for edge model output.
func EdgeInferHTTPResponse(text, model string, latencyMs int64) InferHTTPResponse {
	return InferHTTPResponse{
		Text:      text,
		Model:     model,
		LatencyMs: latencyMs,
		Source:    SourceEdge,
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
		FallbackReason: reason,
		FallbackDetail: r.FallbackDetail,
	}
}
