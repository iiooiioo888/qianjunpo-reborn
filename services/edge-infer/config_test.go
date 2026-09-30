package main

import (
	"testing"
)

func TestNewBackendOllama(t *testing.T) {
	t.Setenv("EDGE_INFER_BACKEND", "ollama")
	cfg := loadConfig()
	be, mock := newBackend(cfg)
	if be.Name() != "ollama" || mock != nil {
		t.Fatalf("be=%s mock=%v", be.Name(), mock)
	}
	if cfg.Ollama.BaseURL == "" || cfg.Ollama.Model == "" {
		t.Fatalf("%+v", cfg.Ollama)
	}
}
