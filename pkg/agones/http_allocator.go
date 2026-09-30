package agones

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPAllocator calls the Agones GameServerAllocation API (or compatible proxy).
type HTTPAllocator struct {
	cfg        Config
	HTTPClient *http.Client
}

type allocationResponse struct {
	GameServerName string `json:"gameServerName"`
	Ports          []struct {
		Port int `json:"port"`
	} `json:"ports"`
	Address string `json:"address"`
}

// NewHTTPAllocator builds an allocator from Config.
func NewHTTPAllocator(cfg Config) *HTTPAllocator {
	return &HTTPAllocator{
		cfg: cfg,
		HTTPClient: &http.Client{
			Timeout: cfg.AllocateTimeout,
		},
	}
}

func (a *HTTPAllocator) Allocate(ctx context.Context, req AllocateRequest) (RoomIdentity, error) {
	if strings.TrimSpace(a.cfg.AllocationURL) == "" {
		return RoomIdentity{}, fmt.Errorf("http allocator: ROMA_AGONES_ALLOCATION_URL is empty")
	}
	body := buildAllocationRequest(a.cfg, req)
	raw, err := json.Marshal(body)
	if err != nil {
		return RoomIdentity{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.cfg.AllocationURL, bytes.NewReader(raw))
	if err != nil {
		return RoomIdentity{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if a.cfg.AllocationToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+a.cfg.AllocationToken)
	}
	resp, err := a.HTTPClient.Do(httpReq)
	if err != nil {
		return RoomIdentity{}, err
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return RoomIdentity{}, fmt.Errorf("allocation HTTP %s: %s", resp.Status, strings.TrimSpace(string(payload)))
	}
	var out allocationResponse
	if err := json.Unmarshal(payload, &out); err != nil {
		return RoomIdentity{}, err
	}
	port := 0
	if len(out.Ports) > 0 {
		port = out.Ports[0].Port
	}
	return RoomIdentity{
		Name:      out.GameServerName,
		Namespace: a.cfg.Namespace,
		Address:   out.Address,
		Port:      port,
		RoomID:    fmt.Sprintf("room-%s-%d", req.ZoneID, req.Shard),
	}, nil
}
