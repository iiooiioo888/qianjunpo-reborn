package ai

import (
	"encoding/json"
	"testing"
)

func TestEdgeInferHTTPResponseJSON(t *testing.T) {
	raw, err := json.Marshal(EdgeInferHTTPResponse("advance: gate", "qwen2.5-3b-mock", 3))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["source"] != SourceEdge {
		t.Fatalf("source=%v", m["source"])
	}
	if _, ok := m["fallback_reason"]; ok {
		t.Fatalf("fallback_reason should be omitted: %v", m)
	}
	if _, ok := m["fallback_detail"]; ok {
		t.Fatalf("fallback_detail should be omitted: %v", m)
	}
}

func TestInferResultFromHTTPBackwardCompat(t *testing.T) {
	out := InferResultFromHTTP(InferHTTPResponse{Text: "ok", Model: "m", LatencyMs: 1})
	if out.Source != SourceEdge || out.FallbackReason != FallbackReasonNone || out.Text != "ok" {
		t.Fatalf("%+v", out)
	}
	if out.RAGK != 0 || len(out.RAGHitIDs) != 0 {
		t.Fatalf("rag fields should be empty: %+v", out)
	}
}

func TestInferResultFromHTTPRAGFields(t *testing.T) {
	out := InferResultFromHTTP(InferHTTPResponse{
		Text:      "line",
		Model:     "m",
		LatencyMs: 2,
		Source:    SourceEdge,
		RAGK:      5,
		RAGHitIDs: []string{"tactic-flank", "wei-cao"},
	})
	if out.RAGK != 5 || len(out.RAGHitIDs) != 2 || out.RAGHitIDs[0] != "tactic-flank" {
		t.Fatalf("%+v", out)
	}
}

func TestInferResultFromHTTPNPCOnWire(t *testing.T) {
	out := InferResultFromHTTP(InferHTTPResponse{
		Text:           "Hold the line!",
		Source:         SourceNPC,
		FallbackReason: string(FallbackReasonTimeout),
		FallbackDetail: "edge proxy timeout",
	})
	if out.Source != SourceNPC || out.FallbackReason != FallbackReasonTimeout || out.FallbackDetail != "edge proxy timeout" {
		t.Fatalf("%+v", out)
	}
}
