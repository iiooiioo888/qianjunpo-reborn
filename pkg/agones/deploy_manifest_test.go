package agones

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func agonesDeployDir(t *testing.T) string {
	return filepath.Join(repoRoot(t), "deploy", "agones")
}

func readDeployFile(t *testing.T, name string) string {
	raw, err := os.ReadFile(filepath.Join(agonesDeployDir(t), name))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func mustContain(t *testing.T, file, body, needle string) {
	t.Helper()
	if !strings.Contains(body, needle) {
		t.Fatalf("%s: expected substring %q", file, needle)
	}
}

func TestAgonesDeployManifestsFleet(t *testing.T) {
	const file = "fleet.yaml"
	body := readDeployFile(t, file)
	mustContain(t, file, body, "apiVersion: agones.dev/v1")
	mustContain(t, file, body, "kind: Fleet")
	mustContain(t, file, body, "name: roma-fleet")
	mustContain(t, file, body, "counters:")
	mustContain(t, file, body, "players:")
	mustContain(t, file, body, "capacity: 64")
	mustContain(t, file, body, "ROMA_AGONES_BACKEND")
	mustContain(t, file, body, "1.54")
}

func TestAgonesDeployManifestsBufferAutoscaler(t *testing.T) {
	const file = "buffer-autoscaler.yaml"
	body := readDeployFile(t, file)
	mustContain(t, file, body, "apiVersion: autoscaling.agones.dev/v1")
	mustContain(t, file, body, "kind: FleetAutoscaler")
	mustContain(t, file, body, "name: roma-buffer-autoscaler")
	mustContain(t, file, body, "fleetName: roma-fleet")
	mustContain(t, file, body, "type: Buffer")
	mustContain(t, file, body, "bufferSize: 5")
	mustContain(t, file, body, "minReplicas: 2")
	mustContain(t, file, body, "maxReplicas: 20")
}

func TestAgonesDeployManifestsCounterAutoscaler(t *testing.T) {
	const file = "fleetautoscaler-counter.yaml"
	body := readDeployFile(t, file)
	mustContain(t, file, body, "type: Counter")
	mustContain(t, file, body, "key: players")
	mustContain(t, file, body, "bufferSize: 32")
	mustContain(t, file, body, "maxCapacity: 2048")
}

func TestAgonesDeployManifestsAllocationPlayers(t *testing.T) {
	const file = "gameserverallocation-players.yaml"
	body := readDeployFile(t, file)
	mustContain(t, file, body, "apiVersion: allocation.agones.dev/v1")
	mustContain(t, file, body, "kind: GameServerAllocation")
	mustContain(t, file, body, "agones.dev/fleet: roma-fleet")
	mustContain(t, file, body, "minAvailable: 1")
	mustContain(t, file, body, "action: Increment")
}

func TestAgonesDeployManifestsCounterIndex(t *testing.T) {
	const file = "counter.yaml"
	body := readDeployFile(t, file)
	mustContain(t, file, body, "players")
	mustContain(t, file, body, "fleet.yaml")
	mustContain(t, file, body, "1.54")
}

func TestAgonesDeployFleetNameMatchesConfigDefault(t *testing.T) {
	cfg := LoadConfig()
	body := readDeployFile(t, "fleet.yaml")
	if cfg.FleetName != "roma-fleet" {
		t.Fatalf("unexpected default fleet %q", cfg.FleetName)
	}
	mustContain(t, "fleet.yaml", body, "name: "+cfg.FleetName)
}
