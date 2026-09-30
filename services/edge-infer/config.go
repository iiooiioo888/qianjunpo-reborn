package main

import (
	"os"
	"strings"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/services/edge-infer/backend"
)

// Config is loaded from environment (EDGE_INFER_*).
type Config struct {
	Addr    string
	Backend string
	Ollama  OllamaConfig
}

type OllamaConfig struct {
	BaseURL            string
	HealthProbeTimeout time.Duration
	Model              string
}

func loadConfig() Config {
	return Config{
		Addr:    env("EDGE_INFER_ADDR", ":8088"),
		Backend: strings.ToLower(env("EDGE_INFER_BACKEND", "mock")),
		Ollama: OllamaConfig{
			BaseURL:            env("EDGE_INFER_OLLAMA_URL", "http://127.0.0.1:11434"),
			HealthProbeTimeout: durationEnv("EDGE_INFER_HEALTH_PROBE_TIMEOUT", backend.DefaultHealthProbeTimeout),
			Model:              env("EDGE_INFER_OLLAMA_MODEL", "qwen2.5:3b"),
		},
	}
}

func newBackend(cfg Config) (backend.Backend, *backend.MockBackend) {
	switch cfg.Backend {
	case "ollama":
		return &backend.OllamaBackend{
			BaseURL:            cfg.Ollama.BaseURL,
			ModelName:          cfg.Ollama.Model,
			HealthProbeTimeout: cfg.Ollama.HealthProbeTimeout,
		}, nil
	default:
		m := &backend.MockBackend{}
		return m, m
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func durationEnv(k string, def time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return def
	}
	return d
}
