package backend

import (
	"context"
	"sync/atomic"
)

const MockModelID = "qwen2.5-3b-mock"

// MockBackend is the CI/default deterministic Qwen stub.
type MockBackend struct {
	Heavy atomic.Bool
	Clock Clock
}

func (m *MockBackend) Name() string  { return "mock" }
func (m *MockBackend) Model() string { return MockModelID }

func (m *MockBackend) Health(_ context.Context) Health {
	return Health{
		Status:  "ok",
		Model:   MockModelID,
		Backend: "mock",
		Ready:   true,
	}
}

func (m *MockBackend) Infer(_ context.Context, req InferRequest) (InferResponse, error) {
	clk := m.Clock
	if clk == nil {
		clk = realClock{}
	}
	start := clk.Now()
	text := mockQwenText(req.Prompt, m.Heavy.Load())
	return InferResponse{
		Text:      text,
		Model:     MockModelID,
		LatencyMs: clk.Since(start).Milliseconds(),
	}, nil
}

func mockQwenText(prompt string, heavy bool) string {
	if prompt == "" {
		return "hold position"
	}
	if heavy {
		return "flank east with cavalry: " + prompt
	}
	return "advance: " + prompt
}
