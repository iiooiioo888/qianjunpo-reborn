package agones

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is loaded from ROMA_AGONES_* environment variables.
type Config struct {
	// Backend drives GameServer-side Ready / Shutdown / Health (mock | sidecar).
	Backend string
	// AllocatorBackend drives Allocate (mock | http).
	AllocatorBackend string
	SidecarBaseURL   string
	AllocationURL    string
	AllocationToken  string
	Namespace        string
	FleetName        string
	AllocateTimeout  time.Duration
	ReadyTimeout     time.Duration
	ShutdownTimeout  time.Duration
	HealthInterval   time.Duration
	// GameServerPort is advertised on mock allocate responses (Roma gRPC).
	GameServerPort int
}

const (
	BackendMock    = "mock"
	BackendSidecar = "sidecar"
	AllocatorMock  = "mock"
	AllocatorHTTP  = "http"
)

// LoadConfig reads environment with CI-friendly defaults (mock everywhere).
func LoadConfig() Config {
	return Config{
		Backend:          strings.ToLower(env("ROMA_AGONES_BACKEND", BackendMock)),
		AllocatorBackend: strings.ToLower(env("ROMA_AGONES_ALLOCATOR", AllocatorMock)),
		SidecarBaseURL:   env("ROMA_AGONES_SIDECAR_URL", "http://127.0.0.1:9358"),
		AllocationURL:    env("ROMA_AGONES_ALLOCATION_URL", ""),
		AllocationToken:  env("ROMA_AGONES_ALLOCATION_TOKEN", ""),
		Namespace:        env("ROMA_AGONES_NAMESPACE", "default"),
		FleetName:        env("ROMA_AGONES_FLEET", "roma-fleet"),
		AllocateTimeout:  durationEnv("ROMA_AGONES_ALLOCATE_TIMEOUT", 30*time.Second),
		ReadyTimeout:     durationEnv("ROMA_AGONES_READY_TIMEOUT", 15*time.Second),
		ShutdownTimeout:  durationEnv("ROMA_AGONES_SHUTDOWN_TIMEOUT", 30*time.Second),
		HealthInterval:   durationEnv("ROMA_AGONES_HEALTH_INTERVAL", 2*time.Second),
		GameServerPort:   intEnv("ROMA_AGONES_GAMESERVER_PORT", 9092),
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
	if err != nil {
		return def
	}
	return d
}

func intEnv(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
