package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handleHealth(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code)
	}
}

func TestInferMock(t *testing.T) {
	body := bytes.NewBufferString(`{"prompt":"defend gate"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/infer", body)
	rr := httptest.NewRecorder()
	handleInfer(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code)
	}
	var resp inferResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Model != "qwen2.5-3b-mock" || resp.Text == "" {
		t.Fatalf("%+v", resp)
	}
}
