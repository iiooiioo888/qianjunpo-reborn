package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/services/edge-infer/backend"
)

func TestHealthMock(t *testing.T) {
	m := &backend.MockBackend{}
	srv := &server{backend: m, mock: m}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	srv.handleHealth(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code)
	}
	var h backend.Health
	if err := json.NewDecoder(rr.Body).Decode(&h); err != nil {
		t.Fatal(err)
	}
	if h.Backend != "mock" || h.Model != backend.MockModelID || !h.Ready {
		t.Fatalf("%+v", h)
	}
}

func TestInferMock(t *testing.T) {
	m := &backend.MockBackend{}
	srv := &server{backend: m, mock: m}
	body := bytes.NewBufferString(`{"prompt":"defend gate"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/infer", body)
	rr := httptest.NewRecorder()
	srv.handleInfer(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code)
	}
	var resp inferResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Model != backend.MockModelID || resp.Text == "" {
		t.Fatalf("%+v", resp)
	}
}

func TestNewBackendFromEnv(t *testing.T) {
	t.Setenv("EDGE_INFER_BACKEND", "mock")
	cfg := loadConfig()
	be, mock := newBackend(cfg)
	if be.Name() != "mock" || mock == nil {
		t.Fatalf("be=%s mock=%v", be.Name(), mock)
	}
}
