package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

type inferFailBackend struct{}

func (inferFailBackend) Name() string  { return "fail" }
func (inferFailBackend) Model() string { return "fail" }
func (inferFailBackend) Health(context.Context) backend.Health {
	return backend.Health{Status: "degraded", Backend: "fail", Ready: false}
}
func (inferFailBackend) Infer(context.Context, backend.InferRequest) (backend.InferResponse, error) {
	return backend.InferResponse{}, errors.New("backend unavailable")
}

func TestInferBackendErrorJSON(t *testing.T) {
	srv := &server{backend: inferFailBackend{}}
	body := bytes.NewBufferString(`{"prompt":"defend gate"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/infer", body)
	rr := httptest.NewRecorder()
	srv.handleInfer(rr, req)
	if rr.Code != http.StatusBadGateway {
		t.Fatalf("code=%d", rr.Code)
	}
	var errResp inferErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
		t.Fatal(err)
	}
	if errResp.Code != "backend_infer_failed" || errResp.Error == "" {
		t.Fatalf("%+v", errResp)
	}
}

func TestHealthOllamaDegradedViaHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			<-r.Context().Done()
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	be := &backend.OllamaBackend{
		BaseURL:            srv.URL,
		ModelName:          "qwen2.5:3b",
		HealthProbeTimeout: 100 * time.Millisecond,
	}
	httpSrv := &server{backend: be}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	start := time.Now()
	httpSrv.handleHealth(rr, req)
	elapsed := time.Since(start)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code)
	}
	var h backend.Health
	if err := json.NewDecoder(rr.Body).Decode(&h); err != nil {
		t.Fatal(err)
	}
	if h.Ready || h.Status != "degraded" || h.Backend != "ollama" {
		t.Fatalf("%+v", h)
	}
	if elapsed > 3*time.Second {
		t.Fatalf("handleHealth blocked too long: %v", elapsed)
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
