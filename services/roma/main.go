package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	romav1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/roma/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/roma"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/observability/metrics"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timedilation"
	"google.golang.org/grpc"
)

type romaServer struct {
	romav1.UnimplementedRomaZoneServer
	store *roma.Store
}

func (s *romaServer) JoinZone(_ context.Context, req *romav1.JoinZoneRequest) (*romav1.JoinZoneResponse, error) {
	z := req.GetZone()
	b, err := s.store.Join(z.GetZoneId(), z.GetShard())
	if err != nil {
		return nil, err
	}
	return &romav1.JoinZoneResponse{
		BattleId: string(b.ID),
		SimTime:  &commonv1.DualTimestamp{WallUnixMs: b.SimTime.WallUnixMs, SimTick: b.SimTime.SimTick},
	}, nil
}

func (s *romaServer) GetBattleState(_ context.Context, req *romav1.GetBattleStateRequest) (*romav1.BattleState, error) {
	b, err := s.store.Get(roma.BattleID(req.GetBattleId()))
	if err != nil {
		return nil, err
	}
	return &romav1.BattleState{
		BattleId:  string(b.ID),
		Zone:      &commonv1.ZoneRef{ZoneId: b.ZoneID, Shard: b.Shard},
		StateHash: b.StateHash(),
		UnitCount: int32(len(b.Units)),
	}, nil
}

func (s *romaServer) SubmitCommand(_ context.Context, req *romav1.SubmitCommandRequest) (*romav1.SubmitCommandResponse, error) {
	h, err := s.store.SubmitCommand(roma.BattleID(req.GetBattleId()), req.GetPlayerId(), req.GetMoveX(), req.GetMoveY())
	if err != nil {
		return &romav1.SubmitCommandResponse{Accepted: false}, nil
	}
	return &romav1.SubmitCommandResponse{Accepted: true, StateHash: h}, nil
}

func main() {
	grpcAddr := env("ROMA_GRPC_ADDR", ":9092")
	httpAddr := env("ROMA_HTTP_ADDR", ":8092")
	store := roma.NewStore(nil)

	go func() {
		metrics.Register(nil)
		metrics.SetTimeFlowRate(int64(timedilation.MaxRate))
		metrics.BattleLatencyP99.Set(12)
		metrics.QueueLen.Set(0)

		mux := http.NewServeMux()
		mux.Handle("/metrics", metrics.Handler())
		mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok","service":"roma","note":"authoritative state in-memory only"}`))
		})
		s := &http.Server{Addr: httpAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
		log.Fatal(s.ListenAndServe())
	}()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	romav1.RegisterRomaZoneServer(srv, &romaServer{store: store})
	log.Printf("roma grpc on %s", grpcAddr)
	log.Fatal(srv.Serve(lis))
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
