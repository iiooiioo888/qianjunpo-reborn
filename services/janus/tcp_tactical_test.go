package main

import (
	"context"
	"net"
	"testing"
	"time"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	gatewayv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/gateway/v1"
	romav1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/roma/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/janus"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/janus/tcpwire"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/roma"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timesync"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestTCPPacketsTacticalLockstep(t *testing.T) {
	romaLis := bufconn.Listen(1 << 20)
	romaSrv := grpc.NewServer()
	romaStore := roma.NewStore(nil)
	romav1.RegisterRomaZoneServer(romaSrv, newE2ERomaServer(romaStore))
	go func() { _ = romaSrv.Serve(romaLis) }()
	defer romaSrv.Stop()

	romaConn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return romaLis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer romaConn.Close()
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

	client, server := net.Pipe()
	defer client.Close()
	go janus.ServeTCPBridge(server, gw)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tcpwire.WritePacket(client, tcpwire.OpConnect, tcpwire.EncodeConnect("tok", "default", 0)); err != nil {
		t.Fatal(err)
	}
	pkt, err := tcpwire.ReadPacket(client)
	if err != nil || pkt.Header.Opcode != tcpwire.OpConnectAck {
		t.Fatalf("connect ack: %v %v", pkt.Header.Opcode, err)
	}
	sess, _, _, _, err := tcpwire.DecodeConnectAck(pkt.Body)
	if err != nil || sess == "" {
		t.Fatal(sess, err)
	}

	if err := tcpwire.WritePacket(client, tcpwire.OpEnterBattle, tcpwire.EncodeEnterBattle(sess, "tok", "default", 0)); err != nil {
		t.Fatal(err)
	}
	pkt, err = tcpwire.ReadPacket(client)
	if err != nil || pkt.Header.Opcode != tcpwire.OpEnterBattleAck {
		t.Fatal(pkt.Header.Opcode, err)
	}
	battleID, _, err := tcpwire.DecodeEnterBattleAck(pkt.Body)
	if err != nil || battleID == "" {
		t.Fatal(battleID, err)
	}

	ref := tactical.NewMatch(referenceSeed("default", 0))
	sched := tactical.DemoSchedule()
	idx := 0
	for frame := uint64(0); frame < 8; frame++ {
		for idx < len(sched) && sched[idx].SubmitFrame == frame {
			item := sched[idx]
			body := tcpwire.EncodeSubmitTactical(sess, battleID, item.Cmd)
			if err := tcpwire.WritePacket(client, tcpwire.OpSubmitTactical, body); err != nil {
				t.Fatal(err)
			}
			pkt, err = tcpwire.ReadPacket(client)
			if err != nil {
				t.Fatal(err)
			}
			ok, reason, _, _, err := tcpwire.DecodeSubmitTacticalAck(pkt.Body)
			if err != nil || !ok {
				t.Fatalf("submit frame %d: %s err=%v", frame, reason, err)
			}
			_ = ref.Submit(item.Cmd)
			idx++
		}
		if err := tcpwire.WritePacket(client, tcpwire.OpStepLockstep, tcpwire.EncodeStepLockstep(sess, battleID, 1)); err != nil {
			t.Fatal(err)
		}
		pkt, err = tcpwire.ReadPacket(client)
		if err != nil {
			t.Fatal(err)
		}
		_, hash, _, _, err := tcpwire.DecodeStepLockstepAck(pkt.Body)
		if err != nil {
			t.Fatal(err)
		}
		ref.StepLockstep()
		if hash != ref.StateHash() {
			t.Fatalf("frame %d hash mismatch", frame)
		}
	}
	_ = ctx
	_ = gatewayv1.ConnectRequest{}
	_ = commonv1.ZoneRef{}
}
