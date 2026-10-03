package main

import (
	"context"
	"encoding/json"
	"net"
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

const bufSize = 1 << 20

func TestJanusToRomaTacticalLockstepPath(t *testing.T) {
	romaLis := bufconn.Listen(bufSize)
	romaSrv := grpc.NewServer()
	romaStore := roma.NewStore(nil)
	romav1.RegisterRomaZoneServer(romaSrv, newE2ERomaServer(romaStore))
	go func() { _ = romaSrv.Serve(romaLis) }()
	defer romaSrv.Stop()

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
	defer romaConn.Close()
	romaTarget := "passthrough:///bufnet"

	janusLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	gw := &janusGateway{
		clock: timesync.NewClock(nil),
		limit: janus.NewRateLimiter(10000),
		disco: &janus.Discovery{Static: map[string]string{"default": romaTarget}, Default: romaTarget},
		auth:  janus.StaticAuth{},
		roma:  janus.NewRomaClient(),
	}
	gw.roma.SetDialer(func(target string) (romav1.RomaZoneClient, error) {
		if target != romaTarget {
			t.Fatalf("unexpected roma target %s", target)
		}
		return romav1.NewRomaZoneClient(romaConn), nil
	})
	janusSrv := grpc.NewServer()
	gatewayv1.RegisterJanusGatewayServer(janusSrv, gw)
	go func() { _ = janusSrv.Serve(janusLis) }()
	defer janusSrv.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	janusConn, err := grpc.DialContext(ctx, janusLis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer janusConn.Close()
	jc := gatewayv1.NewJanusGatewayClient(janusConn)

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
		t.Fatal("expected battle id and initial hash")
	}
	if len(enter.GetViewSnapshotJson()) == 0 {
		t.Fatal("expected enter battle view snapshot json")
	}
	ref := tactical.NewMatch(referenceSeed("default", 0))
	assertViewSnapshotMatchesRef(t, enter.GetViewSnapshotJson(), ref)
	sched := tactical.DemoSchedule()
	idx := 0
	for frame := uint64(0); frame < tactical.DemoTargetFrame(); frame++ {
		for idx < len(sched) && sched[idx].SubmitFrame == frame {
			item := sched[idx]
			sub, err := jc.SubmitTacticalCommand(ctx, &gatewayv1.SubmitTacticalCommandRequest{
				SessionId: connect.GetSessionId(),
				BattleId:  battleID,
				PlayerId:  uint32(item.Cmd.PlayerID),
				Kind:      uint32(item.Cmd.Kind),
				UnitId:    item.Cmd.UnitID,
				ToX:       int32(item.Cmd.To.X),
				ToY:       int32(item.Cmd.To.Y),
			})
			if err != nil {
				t.Fatal(err)
			}
			if !sub.GetAccepted() {
				t.Fatalf("rejected at frame %d: %s", frame, sub.GetRejectReason())
			}
			if err := ref.Submit(item.Cmd); err != nil {
				t.Fatalf("ref submit: %v", err)
			}
			idx++
		}
		st, err := jc.StepTacticalLockstep(ctx, &gatewayv1.StepTacticalLockstepRequest{
			SessionId: connect.GetSessionId(),
			BattleId:  battleID,
			Steps:     1,
		})
		if err != nil {
			t.Fatal(err)
		}
		ref.StepLockstep()
		if st.GetStateHash() != ref.StateHash() {
			t.Fatalf("frame %d hash mismatch svc=%016x ref=%016x", frame, st.GetStateHash(), ref.StateHash())
		}
		if len(st.GetViewSnapshotJson()) == 0 {
			t.Fatalf("frame %d missing view snapshot", frame)
		}
		assertViewSnapshotMatchesRef(t, st.GetViewSnapshotJson(), ref)
	}
	if ref.StateHash() == 0 {
		t.Fatal("expected non-zero ref hash")
	}

	snap, err := jc.GetBattleSnapshot(ctx, &gatewayv1.GetBattleSnapshotRequest{
		SessionId: connect.GetSessionId(),
		BattleId:  battleID,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertViewSnapshotMatchesRef(t, snap.GetViewSnapshotJson(), ref)
}

func assertViewSnapshotMatchesRef(t *testing.T, raw []byte, ref *tactical.Match) {
	var view tactical.ViewSnapshot
	if err := json.Unmarshal(raw, &view); err != nil {
		t.Fatalf("snapshot json: %v", err)
	}
	want := tactical.MatchToViewSnapshot(ref, view.TimeFlowRateParts)
	if view.LockstepFrame != want.LockstepFrame {
		t.Fatalf("snapshot frame=%d ref=%d", view.LockstepFrame, want.LockstepFrame)
	}
	if view.InitialStateHash != want.InitialStateHash {
		t.Fatalf("snapshot initial hash=%s want=%s", view.InitialStateHash, want.InitialStateHash)
	}
	byID := make(map[uint32]tactical.ViewUnit, len(want.Units))
	for _, u := range want.Units {
		byID[u.ID] = u
	}
	for _, u := range view.Units {
		w, ok := byID[u.ID]
		if !ok {
			t.Fatalf("unexpected unit id %d", u.ID)
		}
		if u.X != w.X || u.Y != w.Y || u.HP != w.HP {
			t.Fatalf("unit %d: got %+v want %+v", u.ID, u, w)
		}
	}
}

type e2eRomaServer struct {
	romav1.UnimplementedRomaZoneServer
	store *roma.Store
}

func newE2ERomaServer(store *roma.Store) *e2eRomaServer {
	return &e2eRomaServer{store: store}
}

func (s *e2eRomaServer) JoinZone(_ context.Context, req *romav1.JoinZoneRequest) (*romav1.JoinZoneResponse, error) {
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

func (s *e2eRomaServer) GetBattleState(_ context.Context, req *romav1.GetBattleStateRequest) (*romav1.BattleState, error) {
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

func (s *e2eRomaServer) SubmitTacticalCommand(_ context.Context, req *romav1.SubmitTacticalCommandRequest) (*romav1.SubmitTacticalCommandResponse, error) {
	cmd := req.GetCommand()
	frame, hash, err := s.store.SubmitTacticalCommand(
		roma.BattleID(req.GetBattleId()),
		cmd.GetPlayerId(),
		cmd.GetKind(),
		cmd.GetUnitId(),
		cmd.GetToX(),
		cmd.GetToY(),
		cmd.GetSkillId(),
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

func (s *e2eRomaServer) StepLockstep(_ context.Context, req *romav1.StepLockstepRequest) (*romav1.StepLockstepResponse, error) {
	frame, hash, finished, winner, autoWire, err := s.store.StepLockstep(roma.BattleID(req.GetBattleId()), req.GetSteps(), roma.StepLockstepOpts{
		OneShotAuto:     req.GetAutoCommand(),
		SetAutoModeWire: req.GetAutoCommandMode(),
	})
	if err != nil {
		return nil, err
	}
	return &romav1.StepLockstepResponse{
		LockstepFrame:   frame,
		StateHash:       hash,
		Finished:        finished,
		Winner:          winner,
		AutoCommandMode: autoWire,
	}, nil
}

func (s *e2eRomaServer) GetTacticalViewSnapshot(_ context.Context, req *romav1.GetTacticalViewSnapshotRequest) (*romav1.GetTacticalViewSnapshotResponse, error) {
	jsonBytes, hash, frame, err := s.store.TacticalViewSnapshotJSON(roma.BattleID(req.GetBattleId()))
	if err != nil {
		return nil, err
	}
	return &romav1.GetTacticalViewSnapshotResponse{
		ViewSnapshotJson: jsonBytes,
		StateHash:        hash,
		LockstepFrame:    frame,
	}, nil
}

func referenceSeed(zoneID string, shard uint32) uint64 {
	store := roma.NewStore(nil)
	b, err := store.Join(zoneID, shard)
	if err != nil {
		panic(err)
	}
	return b.Match.Seed
}
