package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OllamaBackend calls an OpenAI-compatible chat API (Ollama default :11434/v1).
type OllamaBackend struct {
	BaseURL    string
	ModelName  string
	HTTPClient *http.Client
	Clock      Clock
}

func (o *OllamaBackend) Name() string  { return "ollama" }
func (o *OllamaBackend) Model() string { return o.ModelName }

func (o *OllamaBackend) client() *http.Client {
	if o.HTTPClient != nil {
		return o.HTTPClient
	}
	return &http.Client{Timeout: 120 * time.Second}
}

func (o *OllamaBackend) clock() Clock {
	if o.Clock != nil {
		return o.Clock
	}
	return realClock{}
}

func (o *OllamaBackend) Health(ctx context.Context) Health {
	h := Health{
		Status:  "ok",
		Model:   o.ModelName,
		Backend: "ollama",
		Ready:   false,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(o.BaseURL, "/")+"/api/tags", nil)
	if err != nil {
		h.Status = "degraded"
		h.Detail = err.Error()
		return h
	}
	resp, err := o.client().Do(req)
	if err != nil {
		h.Status = "degraded"
		h.Detail = "ollama unreachable: " + err.Error()
		return h
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		h.Status = "degraded"
		h.Detail = fmt.Sprintf("ollama tags HTTP %d", resp.StatusCode)
		return h
	}
	h.Ready = true
	return h
}

func (o *OllamaBackend) Infer(ctx context.Context, req InferRequest) (InferResponse, error) {
	clk := o.clock()
	start := clk.Now()
	url := strings.TrimRight(o.BaseURL, "/") + "/v1/chat/completions"
	payload := map[string]interface{}{
		"model": o.ModelName,
		"messages": []map[string]string{
			{"role": "user", "content": req.Prompt},
		},
		"stream": false,
	}
	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return InferResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := o.client().Do(httpReq)
	if err != nil {
		return InferResponse{}, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return InferResponse{}, fmt.Errorf("ollama infer HTTP %d: %s", resp.StatusCode, truncate(raw, 256))
	}
	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return InferResponse{}, err
	}
	text := ""
	if len(decoded.Choices) > 0 {
		text = strings.TrimSpace(decoded.Choices[0].Message.Content)
	}
	return InferResponse{
		Text:      text,
		Model:     o.ModelName,
		LatencyMs: clk.Since(start).Milliseconds(),
	}, nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
