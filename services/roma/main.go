package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	romav1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/roma/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/roma"
	etcdreg "github.com/iiooiioo888/qianjunpo-reborn/pkg/discovery/etcd"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/observability/metrics"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/replay"
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
		UnitCount: b.UnitCount(),
	}, nil
}

func (s *romaServer) SubmitCommand(_ context.Context, req *romav1.SubmitCommandRequest) (*romav1.SubmitCommandResponse, error) {
	h, err := s.store.SubmitCommand(roma.BattleID(req.GetBattleId()), req.GetPlayerId(), req.GetMoveX(), req.GetMoveY())
	if err != nil {
		return &romav1.SubmitCommandResponse{Accepted: false}, nil
	}
	return &romav1.SubmitCommandResponse{Accepted: true, StateHash: h}, nil
}

func (s *romaServer) SubmitTacticalCommand(_ context.Context, req *romav1.SubmitTacticalCommandRequest) (*romav1.SubmitTacticalCommandResponse, error) {
	cmd := req.GetCommand()
	frame, hash, err := s.store.SubmitTacticalCommand(
		roma.BattleID(req.GetBattleId()),
		cmd.GetPlayerId(),
		cmd.GetKind(),
		cmd.GetUnitId(),
		cmd.GetToX(),
		cmd.GetToY(),
	)
	if err != nil {
		return &romav1.SubmitTacticalCommandResponse{
			Accepted:      false,
			RejectReason:  err.Error(),
			StateHash:     hash,
			LockstepFrame: frame,
		}, nil
	}
	return &romav1.SubmitTacticalCommandResponse{
		Accepted:      true,
		StateHash:     hash,
		LockstepFrame: frame,
	}, nil
}

func (s *romaServer) StepLockstep(_ context.Context, req *romav1.StepLockstepRequest) (*romav1.StepLockstepResponse, error) {
	frame, hash, finished, winner, err := s.store.StepLockstep(roma.BattleID(req.GetBattleId()), req.GetSteps())
	if err != nil {
		return nil, err
	}
	return &romav1.StepLockstepResponse{
		LockstepFrame: frame,
		StateHash:     hash,
		Finished:      finished,
		Winner:        winner,
	}, nil
}

func main() {
	grpcAddr := env("ROMA_GRPC_ADDR", ":9092")
	httpAddr := env("ROMA_HTTP_ADDR", ":8092")
	etcdEndpoints := env("ETCD_ENDPOINTS", "etcd:2379")
	advertise := env("ROMA_ADVERTISE_ADDR", "roma:9092")
	zoneID := env("ROMA_ZONE_ID", "default")
	shard := uint32(0)
	if v := env("ROMA_SHARD", "0"); v != "" {
		if n, err := parseShard(v); err == nil {
			shard = n
		}
	}

	store := roma.NewStore(nil)

	if strings.TrimSpace(etcdEndpoints) != "" && os.Getenv("ETCD_DISABLE") != "1" {
		reg, err := etcdreg.NewRomaRegistry(etcdEndpoints)
		if err != nil {
			log.Printf("roma: etcd client failed: %v", err)
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := reg.Register(ctx, zoneID, shard, advertise); err != nil {
				log.Printf("roma: etcd register failed: %v", err)
			} else {
				log.Printf("roma: registered %s shard %d -> %s", zoneID, shard, advertise)
			}
			cancel()
			defer func() { _ = reg.Close() }()
		}
	}

	go func() {
		metrics.Register(nil)
		metrics.SetTimeFlowRate(int64(timedilation.MaxRate))
		metrics.BattleLatencyP99.Set(12)
		metrics.QueueLen.Set(0)

		mux := http.NewServeMux()
		mux.Handle("/metrics", metrics.Handler())
		mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok","service":"roma","note":"authoritative tactical match in-memory"}`))
		})
		mux.HandleFunc("/v1/battles/replay", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			id := r.URL.Query().Get("battle_id")
			if id == "" {
				http.Error(w, "missing battle_id", http.StatusBadRequest)
				return
			}
			rec, err := store.ExportRecording(roma.BattleID(id))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			gz, err := replay.MarshalGzip(rec)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/gzip")
			w.Header().Set("Content-Disposition", "attachment; filename=match.rgz")
			_, _ = w.Write(gz)
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

func parseShard(v string) (uint32, error) {
	var n uint64
	for _, c := range v {
		if c < '0' || c > '9' {
			return 0, os.ErrInvalid
		}
		n = n*10 + uint64(c-'0')
	}
	return uint32(n), nil
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
