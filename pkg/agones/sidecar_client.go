package agones

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

// SidecarClient talks to the Agones SDK sidecar HTTP gateway (default :9358).
type SidecarClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewSidecarClient constructs a sidecar REST client.
func NewSidecarClient(baseURL string) *SidecarClient {
	return &SidecarClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type gameServerBody struct {
	ObjectMeta struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"objectMeta"`
	Status struct {
		State string `json:"state"`
		Ports []struct {
			Port int `json:"port"`
		} `json:"ports"`
		Address string `json:"address"`
	} `json:"status"`
}

func (c *SidecarClient) Identity(ctx context.Context) (RoomIdentity, error) {
	gs, err := c.fetchGameServer(ctx)
	if err != nil {
		return RoomIdentity{}, err
	}
	port := 0
	if len(gs.Status.Ports) > 0 {
		port = gs.Status.Ports[0].Port
	}
	return RoomIdentity{
		Name:      gs.ObjectMeta.Name,
		Namespace: gs.ObjectMeta.Namespace,
		Address:   gs.Status.Address,
		Port:      port,
	}, nil
}

func (c *SidecarClient) Ready(ctx context.Context) error {
	return c.post(ctx, "/ready")
}

func (c *SidecarClient) Shutdown(ctx context.Context) error {
	return c.post(ctx, "/shutdown")
}

func (c *SidecarClient) Health(ctx context.Context) error {
	return c.post(ctx, "/health")
}

func (c *SidecarClient) fetchGameServer(ctx context.Context) (*gameServerBody, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/gameserver", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("sidecar GET /gameserver: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var gs gameServerBody
	if err := json.NewDecoder(resp.Body).Decode(&gs); err != nil {
		return nil, err
	}
	return &gs, nil
}

func (c *SidecarClient) post(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(nil))
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("sidecar POST %s: %s: %s", path, resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

// SidecarStatusSnapshot queries sidecar GameServer JSON for observability.
func (c *SidecarClient) SidecarStatusSnapshot(ctx context.Context) (StatusSnapshot, error) {
	gs, err := c.fetchGameServer(ctx)
	if err != nil {
		return StatusSnapshot{}, err
	}
	ident, _ := c.Identity(ctx)
	phase := mapSidecarState(gs.Status.State)
	return StatusSnapshot{
		Backend:          BackendSidecar,
		AllocatorBackend: AllocatorHTTP,
		Phase:            phase,
		Identity:         ident,
		Ready:            phase == PhaseReady,
	}, nil
}

func mapSidecarState(state string) Phase {
	switch strings.ToLower(state) {
	case "ready":
		return PhaseReady
	case "allocated":
		return PhaseAllocated
	case "shutdown":
		return PhaseShutdown
	case "unhealthy":
		return PhaseError
	default:
		return PhaseUninitialized
	}
}
