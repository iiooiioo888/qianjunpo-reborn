package agones_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/agones"
)

func TestSidecarReadyShutdown(t *testing.T) {
	state := "Starting"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/gameserver":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"objectMeta":{"name":"gs-test","namespace":"default"},"status":{"state":"` + state + `","address":"10.0.0.1","ports":[{"port":7777}]}}`))
		case "/ready":
			state = "Ready"
			w.WriteHeader(http.StatusOK)
		case "/shutdown":
			state = "Shutdown"
			w.WriteHeader(http.StatusOK)
		case "/health":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cli := agones.NewSidecarClient(srv.URL)
	ctx := context.Background()
	if err := cli.Ready(ctx); err != nil {
		t.Fatalf("ready: %v", err)
	}
	snap, err := cli.SidecarStatusSnapshot(ctx)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if snap.Phase != agones.PhaseReady {
		t.Fatalf("phase: %+v", snap)
	}
	if err := cli.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}
