package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	gatewayv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/gateway/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/janus"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/observability/metrics"
	qjptrace "github.com/iiooiioo888/qianjunpo-reborn/pkg/observability/trace"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timesync"
	"google.golang.org/grpc"
)

type janusGateway struct {
	gatewayv1.UnimplementedJanusGatewayServer
	clock *timesync.Clock
	limit *janus.RateLimiter
	disco *janus.Discovery
	auth  janus.AuthHook
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
	zone := req.GetTargetZone().GetZoneId()
	if zone == "" {
		zone = "default"
	}
	romaEP, _ := g.disco.Lookup(ctx, zone)
	now := g.clock.Now()
	_ = pid
	return &gatewayv1.ConnectResponse{
		SessionId:    randomSessionID(),
		ServerTime:   &commonv1.DualTimestamp{WallUnixMs: now.WallUnixMs, SimTick: now.SimTick},
		RomaEndpoint: romaEP,
	}, nil
}

func (g *janusGateway) Heartbeat(_ context.Context, req *gatewayv1.HeartbeatRequest) (*gatewayv1.HeartbeatResponse, error) {
	if req.GetSessionId() == "" {
		return nil, fmt.Errorf("janus: missing session")
	}
	now := g.clock.Now()
	return &gatewayv1.HeartbeatResponse{
		ServerTime: &commonv1.DualTimestamp{WallUnixMs: now.WallUnixMs, SimTick: now.SimTick},
	}, nil
}

func main() {
	grpcAddr := env("JANUS_GRPC_ADDR", ":9090")
	tcpAddr := env("JANUS_TCP_ADDR", ":7000")
	httpAddr := env("JANUS_HTTP_ADDR", ":8090")
	etcdEndpoints := env("ETCD_ENDPOINTS", "etcd:2379")

	gw := &janusGateway{
		clock: timesync.NewClock(nil),
		limit: janus.NewRateLimiter(1000),
		disco: &janus.Discovery{Endpoints: map[string]string{"default": "roma:9092"}},
		auth:  janus.StaticAuth{},
	}
	_ = etcdEndpoints // placeholder for future etcd registration

	go serveTCPBridge(tcpAddr, gw)
	go serveHTTP(httpAddr)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	gatewayv1.RegisterJanusGatewayServer(srv, gw)
	log.Printf("janus grpc=%s tcp-bridge=%s etcd=%s", grpcAddr, tcpAddr, etcdEndpoints)
	log.Fatal(srv.Serve(lis))
}

// serveTCPBridge accepts connections and immediately closes after ack (skeleton).
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
		go handleTCPConn(conn, gw)
	}
}

func handleTCPConn(conn net.Conn, gw *janusGateway) {
	defer conn.Close()
	_, _ = conn.Write([]byte("JANUS_OK\n"))
	_, _ = gw.Connect(context.Background(), &gatewayv1.ConnectRequest{AccessToken: "tcp", TargetZone: &commonv1.ZoneRef{ZoneId: "default"}})
	_, _ = io.ReadAll(conn)
}

func serveHTTP(addr string) {
	metrics.Register(nil)
	metrics.OnlinePlayers.Set(1)

	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"janus"}`))
	})
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
