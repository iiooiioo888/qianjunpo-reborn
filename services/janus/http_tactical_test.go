package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
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

func TestHTTPTacticalCommandMirror(t *testing.T) {
	gw, battleID := testJanusGatewayWithRoma(t)

	mux := http.NewServeMux()
	registerTacticalHTTPRoutes(mux, gw, tacticalHTTPOptions{
		CommandMirrorEnabled: func() bool { return true },
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	body := map[string]interface{}{
		"battle_id": battleID,
		"player_id": 0,
		"kind":      uint32(tactical.KindMove),
		"unit_id":   tactical.UnitIDPlayer0,
		"to_x":      5,
		"to_y":      8,
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
		t.Fatalf("expected accepted command, reason=%q", out.RejectReason)
	}
}

func TestHTTPTacticalCommandMirrorDisabled(t *testing.T) {
	gw, battleID := testJanusGatewayWithRoma(t)

	mux := http.NewServeMux()
	registerTacticalHTTPRoutes(mux, gw, tacticalHTTPOptions{
		CommandMirrorEnabled: func() bool { return false },
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	raw := []byte(`{"battle_id":"` + battleID + `","kind":1}`)
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
	gw, _ := testJanusGatewayWithRoma(t)
	mux := http.NewServeMux()
	registerTacticalHTTPRoutes(mux, gw, defaultTacticalHTTPOptions())
	srv := httptest.NewServer(mux)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/v1/tactical/command")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", res.StatusCode)
	}
}

func testJanusGatewayWithRoma(t *testing.T) (*janusGateway, string) {
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
	jc := gatewayv1.NewJanusGatewayClient(janusConn)

	enter, err := jc.EnterBattle(ctx, &gatewayv1.EnterBattleRequest{
		AccessToken: "test-token",
		TargetZone:  &commonv1.ZoneRef{ZoneId: "default", Shard: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	battleID := enter.GetBattleId()
	if battleID == "" {
		t.Fatal("missing battle id")
	}
	return gw, battleID
}
