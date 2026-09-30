package agones

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildAllocationRequestLegacySelectors(t *testing.T) {
	cfg := Config{
		Namespace:                "default",
		FleetName:                "roma-fleet",
		AllocationPlayersCounter: false,
	}
	raw, err := MarshalAllocationRequestJSON(cfg, AllocateRequest{ZoneID: "dev", Shard: 3})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["gameServerSelectors"]; !ok {
		t.Fatalf("expected gameServerSelectors, got %s", raw)
	}
	if _, ok := m["selectors"]; ok {
		t.Fatalf("legacy body must not include selectors: %s", raw)
	}
	if _, ok := m["counters"]; ok {
		t.Fatalf("legacy body must not include counters: %s", raw)
	}
	body := string(raw)
	if !strings.Contains(body, `"agones.dev/fleet":"roma-fleet"`) {
		t.Fatalf("missing fleet label: %s", body)
	}
	if !strings.Contains(body, `"qianjunpo.dev/zone":"dev"`) {
		t.Fatalf("missing zone metadata: %s", body)
	}
}

func TestBuildAllocationRequestPlayersCounter(t *testing.T) {
	cfg := Config{
		Namespace:                     "default",
		FleetName:                     "roma-fleet",
		AllocationPlayersCounter:      true,
		AllocationPlayersMinAvailable: 1,
		AllocationPlayersIncrement:    1,
	}
	raw, err := MarshalAllocationRequestJSON(cfg, AllocateRequest{ZoneID: "z1", Shard: 0})
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	for _, needle := range []string{
		`"scheduling":"Packed"`,
		`"type":"Counter"`,
		`"key":"players"`,
		`"order":"Ascending"`,
		`"gameServerState":"Ready"`,
		`"gameServerState":"Allocated"`,
		`"minAvailable":1`,
		`"action":"Increment"`,
		`"amount":1`,
		`"agones.dev/fleet":"roma-fleet"`,
	} {
		if !strings.Contains(body, needle) {
			t.Fatalf("expected %q in body: %s", needle, body)
		}
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["gameServerSelectors"]; ok {
		t.Fatalf("counter body must not use gameServerSelectors: %s", body)
	}
	if _, ok := m["selectors"]; !ok {
		t.Fatalf("expected selectors: %s", body)
	}
}

func TestHTTPAllocatorAllocateMockServer(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		b, _ := io.ReadAll(r.Body)
		captured = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"gameServerName":"gs-test-1","address":"10.0.0.5","ports":[{"port":7777}]}`))
	}))
	defer srv.Close()

	cfg := Config{
		AllocationURL:                 srv.URL,
		Namespace:                     "default",
		FleetName:                     "roma-fleet",
		AllocationPlayersCounter:      true,
		AllocationPlayersMinAvailable: 1,
		AllocationPlayersIncrement:    1,
	}
	alloc := NewHTTPAllocator(cfg)
	alloc.HTTPClient = srv.Client()

	ident, err := alloc.Allocate(context.Background(), AllocateRequest{ZoneID: "dev", Shard: 1})
	if err != nil {
		t.Fatalf("allocate: %v", err)
	}
	if ident.Name != "gs-test-1" || ident.Port != 7777 || ident.Address != "10.0.0.5" {
		t.Fatalf("unexpected identity: %+v", ident)
	}
	if !strings.Contains(captured, `"action":"Increment"`) {
		t.Fatalf("server did not receive counter body: %s", captured)
	}
}

func TestLoadConfigAllocationPlayersCounterEnv(t *testing.T) {
	t.Setenv("ROMA_AGONES_ALLOCATION_PLAYERS_COUNTER", "true")
	t.Setenv("ROMA_AGONES_ALLOCATION_PLAYERS_MIN_AVAILABLE", "2")
	t.Setenv("ROMA_AGONES_ALLOCATION_PLAYERS_INCREMENT", "3")
	cfg := LoadConfig()
	if !cfg.AllocationPlayersCounter {
		t.Fatal("expected AllocationPlayersCounter true")
	}
	if cfg.AllocationPlayersMinAvailable != 2 || cfg.AllocationPlayersIncrement != 3 {
		t.Fatalf("unexpected counter tuning: min=%d incr=%d", cfg.AllocationPlayersMinAvailable, cfg.AllocationPlayersIncrement)
	}
}
