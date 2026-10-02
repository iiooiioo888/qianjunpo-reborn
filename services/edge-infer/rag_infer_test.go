package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/ai"
	"github.com/iiooiioo888/qianjunpo-reborn/services/edge-infer/backend"
)

func TestRAGInferMockHTTP(t *testing.T) {
	m := &backend.MockBackend{}
	srv := &server{backend: m, mock: m}
	body := bytes.NewBufferString(`{
		"persona":"guard",
		"order":"flank with cavalry",
		"battle_context":{"units_summary":"enemy infantry center","terrain_summary":"hills east"}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/rag-infer", body)
	rr := httptest.NewRecorder()
	srv.handleRAGInfer(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
	var resp ai.InferHTTPResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Text == "" || resp.Model != backend.MockModelID {
		t.Fatalf("%+v", resp)
	}
	if resp.RAGK != 5 {
		t.Fatalf("rag_k=%d", resp.RAGK)
	}
	if len(resp.RAGHitIDs) == 0 {
		t.Fatalf("expected rag_hit_ids, got %+v", resp)
	}
}

func TestRAGInferRequiresOrder(t *testing.T) {
	m := &backend.MockBackend{}
	srv := &server{backend: m, mock: m}
	req := httptest.NewRequest(http.MethodPost, "/v1/rag-infer", bytes.NewBufferString(`{"persona":"guard"}`))
	rr := httptest.NewRecorder()
	srv.handleRAGInfer(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rr.Code)
	}
}
