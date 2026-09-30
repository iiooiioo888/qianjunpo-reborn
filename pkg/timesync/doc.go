// Package timesync implements cross-region Wall/Sim dual timestamps, client–server
// wall offset and RTT estimation (heartbeat-shaped exchanges), lockstep tick alignment,
// zone sim mapping, and cross-zone freeze policy for Phase 4 (whitepaper v6.0).
//
// Simulation ticks remain authoritative for lockstep; wall clock is for I/O boundaries
// and skew estimation only.
package timesync
