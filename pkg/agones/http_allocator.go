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
	URL        string
	Token      string
	Namespace  string
	FleetName  string
	HTTPClient *http.Client
}

type allocationRequest struct {
	Namespace           string              `json:"namespace"`
	GameServerSelectors []allocationSelector `json:"gameServerSelectors,omitempty"`
	Metadata            *allocationMetadata `json:"metadata,omitempty"`
}

type allocationSelector struct {
	MatchLabels map[string]string `json:"matchLabels"`
}

type allocationMetadata struct {
	Labels map[string]string `json:"labels,omitempty"`
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
		URL:       cfg.AllocationURL,
		Token:     cfg.AllocationToken,
		Namespace: cfg.Namespace,
		FleetName: cfg.FleetName,
		HTTPClient: &http.Client{
			Timeout: cfg.AllocateTimeout,
		},
	}
}

func (a *HTTPAllocator) Allocate(ctx context.Context, req AllocateRequest) (RoomIdentity, error) {
	if strings.TrimSpace(a.URL) == "" {
		return RoomIdentity{}, fmt.Errorf("http allocator: ROMA_AGONES_ALLOCATION_URL is empty")
	}
	body := allocationRequest{
		Namespace: a.Namespace,
		GameServerSelectors: []allocationSelector{{
			MatchLabels: map[string]string{
				"agones.dev/fleet": a.FleetName,
			},
		}},
		Metadata: &allocationMetadata{
			Labels: map[string]string{
				"qianjunpo.dev/zone":  req.ZoneID,
				"qianjunpo.dev/shard": fmt.Sprintf("%d", req.Shard),
			},
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return RoomIdentity{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.URL, bytes.NewReader(raw))
	if err != nil {
		return RoomIdentity{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if a.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+a.Token)
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
		Namespace: a.Namespace,
		Address:   out.Address,
		Port:      port,
		RoomID:    fmt.Sprintf("room-%s-%d", req.ZoneID, req.Shard),
	}, nil
}
