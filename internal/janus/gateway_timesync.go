package janus

import (
	"sync"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timesync"
)

// DualTimeFromProto maps gateway/common DualTimestamp to pkg/timesync.DualTime.
func DualTimeFromProto(ts *commonv1.DualTimestamp) timesync.DualTime {
	if ts == nil {
		return timesync.DualTime{}
	}
	return timesync.DualTime{
		WallUnixMs: ts.GetWallUnixMs(),
		SimTick:    ts.GetSimTick(),
	}
}

// DualTimeToProto maps pkg/timesync.DualTime to gateway/common DualTimestamp.
func DualTimeToProto(d timesync.DualTime) *commonv1.DualTimestamp {
	return &commonv1.DualTimestamp{
		WallUnixMs: d.WallUnixMs,
		SimTick:    d.SimTick,
	}
}

// HeartbeatSampleFromGateway builds a HeartbeatSample from gateway-shaped client_time /
// server_time plus the client wall clock when the response was received (t3).
func HeartbeatSampleFromGateway(
	clientTime, serverTime *commonv1.DualTimestamp,
	clientRecvWallMs int64,
) timesync.HeartbeatSample {
	return timesync.NewHeartbeatSample(
		DualTimeFromProto(clientTime),
		DualTimeFromProto(serverTime),
		clientRecvWallMs,
	)
}

// HeartbeatSync wires gateway heartbeat exchanges into WallOffsetEstimator and TickAligner.
type HeartbeatSync struct {
	Estimator *timesync.WallOffsetEstimator
	Aligner   *timesync.TickAligner
}

// NewHeartbeatSync creates sync state with optional sim cadence between heartbeats.
func NewHeartbeatSync(cadence timesync.SimCadence) *HeartbeatSync {
	return &HeartbeatSync{
		Estimator: timesync.NewWallOffsetEstimator(0.25),
		Aligner:   timesync.NewTickAligner(cadence),
	}
}

// ObserveClientRoundTrip ingests a full client-observed heartbeat (preferred for RTT/skew).
func (s *HeartbeatSync) ObserveClientRoundTrip(
	clientTime, serverTime *commonv1.DualTimestamp,
	clientRecvWallMs int64,
) timesync.HeartbeatSample {
	h := HeartbeatSampleFromGateway(clientTime, serverTime, clientRecvWallMs)
	s.ObserveSample(h)
	return h
}

// ObserveServerLeg ingests a Janus Heartbeat RPC on the server: client_time from the request
// and server_time about to be returned. clientRecvWallMs is approximated as server send wall
// (instant downlink); use ObserveClientRoundTrip on the client for accurate RTT.
func (s *HeartbeatSync) ObserveServerLeg(
	clientTime *commonv1.DualTimestamp,
	serverDual timesync.DualTime,
) timesync.HeartbeatSample {
	h := timesync.NewHeartbeatSample(
		DualTimeFromProto(clientTime),
		serverDual,
		serverDual.WallUnixMs,
	)
	s.ObserveSample(h)
	return h
}

// ObserveSample feeds estimators from an already-built sample.
func (s *HeartbeatSync) ObserveSample(h timesync.HeartbeatSample) {
	if s.Estimator != nil {
		s.Estimator.Observe(h)
	}
	if s.Aligner != nil {
		s.Aligner.ObserveHeartbeat(h, s.Estimator)
	}
}

// SessionHeartbeatSync tracks per-session heartbeat sync (Janus session_id).
type SessionHeartbeatSync struct {
	mu   sync.Mutex
	byID map[string]*HeartbeatSync
	cadence timesync.SimCadence
}

// NewSessionHeartbeatSync creates an empty session sync registry.
func NewSessionHeartbeatSync(cadence timesync.SimCadence) *SessionHeartbeatSync {
	return &SessionHeartbeatSync{
		byID:    make(map[string]*HeartbeatSync),
		cadence: cadence,
	}
}

// ForSession returns sync state for sessionID, creating it if needed.
func (r *SessionHeartbeatSync) ForSession(sessionID string) *HeartbeatSync {
	if sessionID == "" {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.byID[sessionID]; ok {
		return s
	}
	s := NewHeartbeatSync(r.cadence)
	r.byID[sessionID] = s
	return s
}
