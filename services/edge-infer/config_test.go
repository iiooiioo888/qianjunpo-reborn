package main

import (
	"testing"
	"time"
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

func TestHealthProbeTimeoutFromEnv(t *testing.T) {
	t.Setenv("EDGE_INFER_HEALTH_PROBE_TIMEOUT", "500ms")
	cfg := loadConfig()
	if cfg.Ollama.HealthProbeTimeout != 500*time.Millisecond {
		t.Fatalf("got %v", cfg.Ollama.HealthProbeTimeout)
	}
}
