package integration

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	gatewayv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/gateway/v1"
	laresv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/lares/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Compose E2E: requires Docker Compose stack (make compose-up) and COMPOSE_E2E=1.
// Cloud Agent VMs often lack Docker; the test skips unless explicitly enabled.
func TestComposeJanusRomaTacticalSequence(t *testing.T) {
	if os.Getenv("COMPOSE_E2E") != "1" {
		t.Skip("set COMPOSE_E2E=1 with `make compose-up` to run compose-level Janus→Roma E2E")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available in this environment")
	}

	janusAddr := envOr("JANUS_GRPC_HOST", "127.0.0.1:19090")
	laresAddr := envOr("LARES_GRPC_HOST", "127.0.0.1:19091")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	laresConn, err := grpc.DialContext(ctx, laresAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial lares: %v", err)
	}
	defer laresConn.Close()
	lc := laresv1.NewLaresAuthClient(laresConn)
	login, err := lc.Login(ctx, &laresv1.LoginRequest{Username: "compose-e2e", Password: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	token := login.GetAccessToken()

	janusConn, err := grpc.DialContext(ctx, janusAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial janus: %v", err)
	}
	defer janusConn.Close()
	jc := gatewayv1.NewJanusGatewayClient(janusConn)

	connect, err := jc.Connect(ctx, &gatewayv1.ConnectRequest{
		AccessToken: token,
		TargetZone:  &commonv1.ZoneRef{ZoneId: "default", Shard: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	enter, err := jc.EnterBattle(ctx, &gatewayv1.EnterBattleRequest{
		SessionId:   connect.GetSessionId(),
		AccessToken: token,
		TargetZone:  &commonv1.ZoneRef{ZoneId: "default", Shard: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	battleID := enter.GetBattleId()

	sched := tactical.DemoSchedule()
	idx := 0
	for frame := uint64(0); frame < 12; frame++ {
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
			if err != nil || !sub.GetAccepted() {
				t.Fatalf("submit frame %d: %v accepted=%v", frame, err, sub.GetAccepted())
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
		if st.GetStateHash() == 0 {
			t.Fatalf("zero hash at frame %d", frame)
		}
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
