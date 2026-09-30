package agones_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/agones"
)

func TestMockAllocateReadyShutdown(t *testing.T) {
	cfg := agones.Config{
		Backend:          agones.BackendMock,
		AllocatorBackend: agones.AllocatorMock,
		GameServerPort:   19092,
		AllocateTimeout:  2 * time.Second,
		ReadyTimeout:     2 * time.Second,
		ShutdownTimeout:  2 * time.Second,
	}
	coord, err := agones.NewCoordinator(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	ident, err := coord.AllocateRoom(ctx, agones.AllocateRequest{ZoneID: "z1", Shard: 2})
	if err != nil {
		t.Fatalf("allocate: %v", err)
	}
	if ident.Name == "" || ident.RoomID != "room-z1-2" {
		t.Fatalf("unexpected identity: %+v", ident)
	}

	if err := coord.BootstrapGameServer(ctx); err != nil {
		t.Fatalf("ready: %v", err)
	}
	st := coord.Status(ctx)
	if !st.Ready || st.Phase != agones.PhaseReady {
		t.Fatalf("status after ready: %+v", st)
	}

	if err := coord.ShutdownGameServer(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	st = coord.Status(ctx)
	if st.Phase != agones.PhaseShutdown {
		t.Fatalf("expected shutdown phase, got %+v", st)
	}
}

func TestReadyTimeoutObservable(t *testing.T) {
	cfg := agones.Config{
		Backend:          agones.BackendMock,
		AllocatorBackend: agones.AllocatorMock,
		ReadyTimeout:     50 * time.Millisecond,
	}
	coord, err := agones.NewCoordinator(cfg)
	if err != nil {
		t.Fatal(err)
	}
	coord.MockBackendForTest().SetReadyDelay(200 * time.Millisecond)

	err = coord.BootstrapGameServer(context.Background())
	var opErr *agones.OperationError
	if !errors.As(err, &opErr) {
		t.Fatalf("expected OperationError, got %v", err)
	}
	if !opErr.Timeout {
		t.Fatalf("expected timeout flag, got %+v", opErr)
	}
	st := coord.Status(context.Background())
	if st.Phase != agones.PhaseError {
		t.Fatalf("expected error phase, got %+v", st)
	}
	if st.LastError == "" {
		t.Fatal("expected last_error populated")
	}
}

func TestShutdownFailureObservable(t *testing.T) {
	cfg := agones.Config{
		Backend:          agones.BackendMock,
		AllocatorBackend: agones.AllocatorMock,
		ShutdownTimeout:  time.Second,
	}
	coord, err := agones.NewCoordinator(cfg)
	if err != nil {
		t.Fatal(err)
	}
	coord.MockBackendForTest().SetFailShutdown(true)
	if err := coord.BootstrapGameServer(context.Background()); err != nil {
		t.Fatalf("ready: %v", err)
	}
	err = coord.ShutdownGameServer(context.Background())
	if err == nil {
		t.Fatal("expected shutdown error")
	}
	st := coord.Status(context.Background())
	if st.LastError == "" {
		t.Fatal("expected last_error")
	}
}
