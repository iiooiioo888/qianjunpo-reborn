package timesync

import "time"

// HeartbeatSample captures one Janus-shaped heartbeat exchange on the client:
// client sends client_time at send wall t0, receives server_time at recv wall t3.
type HeartbeatSample struct {
	ClientSent DualTime
	Server     DualTime
	ClientRecvWallMs int64
}

// NewHeartbeatSample builds a sample from wall millis (tests and gateways).
func NewHeartbeatSample(clientSent, server DualTime, clientRecvWallMs int64) HeartbeatSample {
	return HeartbeatSample{
		ClientSent:       clientSent,
		Server:           server,
		ClientRecvWallMs: clientRecvWallMs,
	}
}

// RTT returns round-trip wall delay; zero if clocks are inconsistent.
func (h HeartbeatSample) RTT() time.Duration {
	if h.ClientRecvWallMs < h.ClientSent.WallUnixMs {
		return 0
	}
	return time.Duration(h.ClientRecvWallMs-h.ClientSent.WallUnixMs) * time.Millisecond
}

// CristianOffsetMs estimates server−client wall skew (ms) assuming symmetric delay.
// Positive means server wall runs ahead of client wall.
func (h HeartbeatSample) CristianOffsetMs() int64 {
	rtt := h.RTT()
	if rtt <= 0 {
		return 0
	}
	half := rtt / 2
	clientMid := h.ClientSent.WallUnixMs + half.Milliseconds()
	return h.Server.WallUnixMs - clientMid
}
