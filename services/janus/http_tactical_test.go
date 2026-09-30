package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	gatewayv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/gateway/v1"
	romav1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/roma/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/janus"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/roma"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timesync"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// janusRomaHTTPFixture wires in-process Janus gRPC + mock Roma for HTTP mirror tests.
type janusRomaHTTPFixture struct {
	GW               *janusGateway
	BattleID         string
	SessionID        string
	InitialStateHash uint64
}

func TestHTTPTacticalCommandMirror(t *testing.T) {
	fix := startJanusGatewayWithEnterBattle(t)
	srv := newTacticalHTTPServer(t, fix.GW, tacticalHTTPOptions{
		CommandMirrorEnabled: func() bool { return true },
	})

	body := map[string]interface{}{
		"session_id": fix.SessionID,
		"battle_id":  fix.BattleID,
		"player_id":  0,
		"kind":       uint32(tactical.KindMove),
		"unit_id":    tactical.UnitIDPlayer0,
		"to_x":       5,
		"to_y":       8,
	}
	raw, _ := json.Marshal(body)
	res, err := http.Post(srv.URL+"/v1/tactical/command", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d: %s", res.StatusCode, b)
	}
	var out tacticalCommandJSON
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if !out.Accepted {
		t.Fatalf("expected accepted command after EnterBattle, reason=%q", out.RejectReason)
	}
	if out.StateHash == 0 {
		t.Fatal("expected non-zero state_hash on accepted command")
	}
	if out.RejectReason != "" {
		t.Fatalf("unexpected reject_reason on success: %q", out.RejectReason)
	}
}

func TestHTTPTacticalCommandMirrorWithoutEnterBattle(t *testing.T) {
	fix := startJanusGatewayRomaOnly(t)
	srv := newTacticalHTTPServer(t, fix.GW, defaultTacticalHTTPOptions())

	// Docs/dev smoke often use default/0; without EnterBattle (JoinZone) Roma has no battle.
	raw := []byte(`{"battle_id":"default/0","player_id":0,"kind":1,"unit_id":1,"to_x":5,"to_y":8}`)
	res, err := http.Post(srv.URL+"/v1/tactical/command", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d: %s", res.StatusCode, b)
	}
	var out tacticalCommandJSON
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Accepted {
		t.Fatal("expected rejected command when battle was never created")
	}
	if !strings.Contains(out.RejectReason, "battle not found") {
		t.Fatalf("expected roma battle not found, got reason=%q", out.RejectReason)
	}
}

func TestHTTPTacticalSnapshotAfterEnterBattle(t *testing.T) {
	fix := startJanusGatewayWithEnterBattle(t)
	srv := newTacticalHTTPServer(t, fix.GW, defaultTacticalHTTPOptions())

	res, err := http.Get(srv.URL + "/v1/tactical/snapshot?battle_id=" + fix.BattleID)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d: %s", res.StatusCode, b)
	}
	if res.Header.Get("X-State-Hash") == "" {
		t.Fatal("missing X-State-Hash header")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) == 0 {
		t.Fatal("empty snapshot body")
	}
	var view tactical.ViewSnapshot
	if err := json.Unmarshal(body, &view); err != nil {
		t.Fatalf("snapshot json: %v", err)
	}
	if view.LockstepFrame != 0 {
		t.Fatalf("expected initial lockstep frame 0, got %d", view.LockstepFrame)
	}
}

func TestHTTPTacticalCommandMirrorDisabled(t *testing.T) {
	fix := startJanusGatewayWithEnterBattle(t)
	srv := newTacticalHTTPServer(t, fix.GW, tacticalHTTPOptions{
		CommandMirrorEnabled: func() bool { return false },
	})

	raw := []byte(`{"battle_id":"` + fix.BattleID + `","kind":1}`)
	res, err := http.Post(srv.URL+"/v1/tactical/command", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", res.StatusCode)
	}
	var out tacticalCommandJSON
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Accepted || out.RejectReason == "" {
		t.Fatalf("expected disabled reject payload, got %+v", out)
	}
}

func TestHTTPTacticalCommandMethodNotAllowed(t *testing.T) {
	fix := startJanusGatewayWithEnterBattle(t)
	srv := newTacticalHTTPServer(t, fix.GW, defaultTacticalHTTPOptions())

	res, err := http.Get(srv.URL + "/v1/tactical/command")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", res.StatusCode)
	}
}

func newTacticalHTTPServer(t *testing.T, gw *janusGateway, opts tacticalHTTPOptions) *httptest.Server {
	mux := http.NewServeMux()
	registerTacticalHTTPRoutes(mux, gw, opts)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// startJanusGatewayRomaOnly stands up Janus + mock Roma without EnterBattle (no live battle).
func startJanusGatewayRomaOnly(t *testing.T) janusRomaHTTPFixture {
	gw, _ := startJanusGRPCWithRoma(t)
	return janusRomaHTTPFixture{GW: gw, BattleID: "default/0"}
}

// startJanusGatewayWithEnterBattle runs Connect → EnterBattle so Roma holds an in-memory battle.
func startJanusGatewayWithEnterBattle(t *testing.T) janusRomaHTTPFixture {
	gw, jc := startJanusGRPCWithRoma(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	connect, err := jc.Connect(ctx, &gatewayv1.ConnectRequest{
		AccessToken: "test-token",
		TargetZone:  &commonv1.ZoneRef{ZoneId: "default", Shard: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	enter, err := jc.EnterBattle(ctx, &gatewayv1.EnterBattleRequest{
		SessionId:   connect.GetSessionId(),
		AccessToken: "test-token",
		TargetZone:  &commonv1.ZoneRef{ZoneId: "default", Shard: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	battleID := enter.GetBattleId()
	if battleID == "" || enter.GetInitialStateHash() == 0 {
		t.Fatal("missing battle id or initial state hash from EnterBattle")
	}
	return janusRomaHTTPFixture{
		GW:               gw,
		BattleID:         battleID,
		SessionID:        connect.GetSessionId(),
		InitialStateHash: enter.GetInitialStateHash(),
	}
}

func startJanusGRPCWithRoma(t *testing.T) (*janusGateway, gatewayv1.JanusGatewayClient) {
	romaLis := bufconn.Listen(bufSize)
	romaSrv := grpc.NewServer()
	romaStore := roma.NewStore(nil)
	romav1.RegisterRomaZoneServer(romaSrv, newE2ERomaServer(romaStore))
	go func() { _ = romaSrv.Serve(romaLis) }()
	t.Cleanup(romaSrv.Stop)

	romaDialer := func(context.Context, string) (net.Conn, error) {
		return romaLis.Dial()
	}
	romaConn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(romaDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = romaConn.Close() })
	romaTarget := "passthrough:///bufnet"

	gw := &janusGateway{
		clock: timesync.NewClock(nil),
		limit: janus.NewRateLimiter(10000),
		disco: &janus.Discovery{Static: map[string]string{"default": romaTarget}, Default: romaTarget},
		auth:  janus.StaticAuth{},
		roma:  janus.NewRomaClient(),
	}
	gw.roma.SetDialer(func(target string) (romav1.RomaZoneClient, error) {
		return romav1.NewRomaZoneClient(romaConn), nil
	})

	janusLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	janusSrv := grpc.NewServer()
	gatewayv1.RegisterJanusGatewayServer(janusSrv, gw)
	go func() { _ = janusSrv.Serve(janusLis) }()
	t.Cleanup(janusSrv.Stop)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	janusConn, err := grpc.DialContext(ctx, janusLis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = janusConn.Close() })
	return gw, gatewayv1.NewJanusGatewayClient(janusConn)
}
