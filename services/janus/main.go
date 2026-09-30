package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	gatewayv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/gateway/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/janus"
	etcdreg "github.com/iiooiioo888/qianjunpo-reborn/pkg/discovery/etcd"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/observability/metrics"
	qjptrace "github.com/iiooiioo888/qianjunpo-reborn/pkg/observability/trace"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timesync"
	"google.golang.org/grpc"
)

type janusGateway struct {
	gatewayv1.UnimplementedJanusGatewayServer
	clock       *timesync.Clock
	heartbeat   *janus.SessionHeartbeatSync
	limit       *janus.RateLimiter
	disco       *janus.Discovery
	auth        janus.AuthHook
	roma        *janus.RomaClient
}

func (g *janusGateway) Connect(ctx context.Context, req *gatewayv1.ConnectRequest) (*gatewayv1.ConnectResponse, error) {
	ctx, span := qjptrace.StartJanusToRomaSpan(ctx, "Connect")
	defer qjptrace.EndSpan(span, nil)

	if !g.limit.Allow() {
		return nil, fmt.Errorf("janus: rate limited")
	}
	pid, ok := g.auth.ValidateAccess(ctx, req.GetAccessToken())
	if !ok {
		return nil, fmt.Errorf("janus: unauthorized")
	}
	zone := janus.NormalizeZoneID(req.GetTargetZone().GetZoneId())
	romaEP, err := g.disco.Lookup(ctx, zone)
	if err != nil {
		log.Printf("janus: Connect zone %q discovery degraded: %v", zone, err)
	}
	now := g.clock.Now()
	sid := randomSessionID()
	if g.heartbeat != nil {
		g.heartbeat.ForSession(sid)
	}
	_ = pid
	return &gatewayv1.ConnectResponse{
		SessionId:    sid,
		ServerTime:   janus.DualTimeToProto(now),
		RomaEndpoint: romaEP,
	}, nil
}

func (g *janusGateway) Heartbeat(_ context.Context, req *gatewayv1.HeartbeatRequest) (*gatewayv1.HeartbeatResponse, error) {
	if req.GetSessionId() == "" {
		return nil, fmt.Errorf("janus: missing session")
	}
	now := g.clock.Now()
	if ct := req.GetClientTime(); ct != nil && g.heartbeat != nil {
		if sync := g.heartbeat.ForSession(req.GetSessionId()); sync != nil {
			sync.ObserveServerLeg(ct, now)
		}
	}
	return &gatewayv1.HeartbeatResponse{
		ServerTime: janus.DualTimeToProto(now),
	}, nil
}

func (g *janusGateway) EnterBattle(ctx context.Context, req *gatewayv1.EnterBattleRequest) (*gatewayv1.EnterBattleResponse, error) {
	ctx, span := qjptrace.StartJanusToRomaSpan(ctx, "EnterBattle")
	var err error
	defer func() { qjptrace.EndSpan(span, err) }()

	if !g.limit.Allow() {
		err = fmt.Errorf("janus: rate limited")
		return nil, err
	}
	if _, ok := g.auth.ValidateAccess(ctx, req.GetAccessToken()); !ok {
		err = fmt.Errorf("janus: unauthorized")
		return nil, err
	}
	zone := req.GetTargetZone()
	if zone == nil || zone.GetZoneId() == "" {
		zone = &commonv1.ZoneRef{ZoneId: "default"}
	}
	romaEP, err := g.disco.Lookup(ctx, janus.NormalizeZoneID(zone.GetZoneId()))
	if err != nil {
		return nil, err
	}
	return g.roma.EnterBattle(ctx, romaEP, req.GetAccessToken(), zone)
}

func (g *janusGateway) SubmitTacticalCommand(ctx context.Context, req *gatewayv1.SubmitTacticalCommandRequest) (*gatewayv1.SubmitTacticalCommandResponse, error) {
	ctx, span := qjptrace.StartJanusToRomaSpan(ctx, "SubmitTacticalCommand")
	var err error
	defer func() { qjptrace.EndSpan(span, err) }()

	if req.GetBattleId() == "" {
		err = fmt.Errorf("janus: missing battle_id")
		return nil, err
	}
	zone := "default"
	romaEP, err := g.disco.Lookup(ctx, zone)
	if err != nil {
		return nil, err
	}
	return g.roma.SubmitTacticalCommand(ctx, romaEP, req)
}

func (g *janusGateway) StepTacticalLockstep(ctx context.Context, req *gatewayv1.StepTacticalLockstepRequest) (*gatewayv1.StepTacticalLockstepResponse, error) {
	ctx, span := qjptrace.StartJanusToRomaSpan(ctx, "StepTacticalLockstep")
	var err error
	defer func() { qjptrace.EndSpan(span, err) }()

	if req.GetBattleId() == "" {
		err = fmt.Errorf("janus: missing battle_id")
		return nil, err
	}
	zone := "default"
	romaEP, err := g.disco.Lookup(ctx, zone)
	if err != nil {
		return nil, err
	}
	return g.roma.StepTacticalLockstep(ctx, romaEP, req)
}

func (g *janusGateway) GetBattleSnapshot(ctx context.Context, req *gatewayv1.GetBattleSnapshotRequest) (*gatewayv1.GetBattleSnapshotResponse, error) {
	ctx, span := qjptrace.StartJanusToRomaSpan(ctx, "GetBattleSnapshot")
	var err error
	defer func() { qjptrace.EndSpan(span, err) }()

	if req.GetBattleId() == "" {
		err = fmt.Errorf("janus: missing battle_id")
		return nil, err
	}
	zone := "default"
	romaEP, err := g.disco.Lookup(ctx, zone)
	if err != nil {
		return nil, err
	}
	return g.roma.GetBattleSnapshot(ctx, romaEP, req)
}

func main() {
	grpcAddr := env("JANUS_GRPC_ADDR", ":9090")
	tcpAddr := env("JANUS_TCP_ADDR", ":7000")
	httpAddr := env("JANUS_HTTP_ADDR", ":8090")
	etcdEndpoints := env("ETCD_ENDPOINTS", "etcd:2379")
	romaDefault := env("ROMA_GRPC_ADDR", "roma:9092")
	laresAddr := env("JANUS_LARES_GRPC_ADDR", "lares:9091")
	laresSecret := env("LARES_TOKEN_SECRET", "phase4-dev-secret")

	staticZones := map[string]string{"default": romaDefault}
	if raw := strings.TrimSpace(os.Getenv("JANUS_ROMA_ZONE_ENDPOINTS")); raw != "" {
		parsed, err := janus.ParseZoneEndpointMap(raw)
		if err != nil {
			log.Fatalf("janus: %v", err)
		}
		for zone, ep := range parsed {
			staticZones[zone] = ep
		}
	}

	disco := &janus.Discovery{
		Default: romaDefault,
		Static:  staticZones,
	}
	var etcdCloser func()
	if strings.TrimSpace(etcdEndpoints) != "" && os.Getenv("ETCD_DISABLE") != "1" {
		reg, err := etcdreg.NewRomaRegistry(etcdEndpoints)
		if err != nil {
			log.Printf("janus: etcd client failed (%v); static roma fallback only", err)
		} else {
			disco.Etcd = &janus.EtcdResolver{Reg: reg}
			etcdCloser = func() { _ = reg.Close() }
		}
	}

	auth, err := janus.NewLaresAuth(laresAddr, laresSecret)
	if err != nil {
		log.Printf("janus: lares auth dial failed (%v); static dev auth fallback", err)
		auth = nil
	}
	var authHook janus.AuthHook
	if auth != nil {
		authHook = auth
	} else {
		authHook = janus.StaticAuth{}
	}

	gw := &janusGateway{
		clock: timesync.NewClock(nil),
		heartbeat: janus.NewSessionHeartbeatSync(timesync.SimCadence{
			TicksPerSecond: envInt64("JANUS_SIM_TICKS_PER_SEC", 10),
		}),
		limit: janus.NewRateLimiter(1000),
		disco: disco,
		auth:  authHook,
		roma:  janus.NewRomaClient(),
	}
	if etcdCloser != nil {
		defer etcdCloser()
	}

	go serveTCPBridge(tcpAddr, gw)
	go serveHTTP(httpAddr, gw)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	gatewayv1.RegisterJanusGatewayServer(srv, gw)
	log.Printf("janus grpc=%s tcp-bridge=%s roma=%s etcd=%s lares=%s", grpcAddr, tcpAddr, romaDefault, etcdEndpoints, laresAddr)
	log.Fatal(srv.Serve(lis))
}

func serveTCPBridge(addr string, gw *janusGateway) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go janus.ServeTCPBridge(conn, gw)
	}
}

func serveHTTP(addr string, gw *janusGateway) {
	metrics.Register(nil)
	metrics.OnlinePlayers.Set(1)

	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"janus"}`))
	})
	registerTacticalHTTPRoutes(mux, gw, defaultTacticalHTTPOptions())
	s := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(s.ListenAndServe())
}

func randomSessionID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt64(k string, def int64) int64 {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}
