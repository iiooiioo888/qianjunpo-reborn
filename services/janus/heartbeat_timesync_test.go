package main

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	gatewayv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/gateway/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/janus"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timesync"
)

func TestGatewayHeartbeatWiresTimesync(t *testing.T) {
	fixed := time.UnixMilli(5_000)
	clk := timesync.NewClock(func() time.Time { return fixed })
	gw := &janusGateway{
		clock:     clk,
		heartbeat: janus.NewSessionHeartbeatSync(timesync.SimCadence{TicksPerSecond: 10}),
		limit:     janus.NewRateLimiter(100),
		auth:      janus.StaticAuth{},
		disco:     &janus.Discovery{Default: "roma:1"},
	}

	connect, err := gw.Connect(context.Background(), &gatewayv1.ConnectRequest{
		AccessToken: "tok",
		TargetZone:  &commonv1.ZoneRef{ZoneId: "default"},
	})
	if err != nil {
		t.Fatal(err)
	}
	sid := connect.GetSessionId()

	_, err = gw.Heartbeat(context.Background(), &gatewayv1.HeartbeatRequest{
		SessionId: sid,
		ClientTime: &commonv1.DualTimestamp{
			WallUnixMs: 4900,
			SimTick:    3,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	sync := gw.heartbeat.ForSession(sid)
	if sync == nil || sync.Estimator.Samples() != 1 {
		t.Fatalf("estimator samples=%d", sync.Estimator.Samples())
	}
}
