package janus

import (
	"testing"
	"time"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timesync"
)

func TestDualTimeProtoRoundTrip(t *testing.T) {
	in := &commonv1.DualTimestamp{WallUnixMs: 42, SimTick: 7}
	out := DualTimeToProto(DualTimeFromProto(in))
	if out.GetWallUnixMs() != 42 || out.GetSimTick() != 7 {
		t.Fatalf("round trip %+v", out)
	}
}

func TestHeartbeatSampleFromGatewayFeedsEstimators(t *testing.T) {
	clientTime := &commonv1.DualTimestamp{WallUnixMs: 1000, SimTick: 5}
	serverTime := &commonv1.DualTimestamp{WallUnixMs: 1050, SimTick: 10}
	sync := NewHeartbeatSync(timesync.SimCadence{TicksPerSecond: 10})

	h := sync.ObserveClientRoundTrip(clientTime, serverTime, 1100)
	if h.RTT() != 100*time.Millisecond {
		t.Fatalf("rtt=%v", h.RTT())
	}
	if sync.Estimator.Samples() != 1 {
		t.Fatalf("samples=%d", sync.Estimator.Samples())
	}
	sim, err := sync.Aligner.EstimatedServerSim(1500)
	if err != nil {
		t.Fatal(err)
	}
	if sim < 10 {
		t.Fatalf("sim=%d", sim)
	}
}

func TestObserveServerLeg(t *testing.T) {
	sync := NewHeartbeatSync(timesync.SimCadence{})
	clientTime := &commonv1.DualTimestamp{WallUnixMs: 2000, SimTick: 1}
	server := timesync.DualTime{WallUnixMs: 2030, SimTick: 2}
	sync.ObserveServerLeg(clientTime, server)
	if sync.Estimator.Samples() != 1 {
		t.Fatal("expected sample")
	}
	sim, err := sync.Aligner.EstimatedServerSim(2030)
	if err != nil || sim != 2 {
		t.Fatalf("sim=%d err=%v", sim, err)
	}
}

func TestSessionHeartbeatSync(t *testing.T) {
	reg := NewSessionHeartbeatSync(timesync.SimCadence{})
	a := reg.ForSession("sess-a")
	b := reg.ForSession("sess-b")
	if a == b {
		t.Fatal("distinct sessions")
	}
	if reg.ForSession("sess-a") != a {
		t.Fatal("cached session")
	}
}
