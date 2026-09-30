// Package loadpredict implements Phase 5 load-forecast scaffolding for production tuning.
//
// Flow (whitepaper v6.0 — 生產調優):
//
//	CPU/queue samples → SeriesPredictor (LSTM pluggable; CI uses fake series)
//	  → 30s-ahead load ≥90% hook
//	  → PreScalePlanner emits Agones buffer/Fleet scale hints 7–15 minutes ahead
//	  → Roma time dilation recovers time_flow_rate after capacity catches up
//
// See deploy/agones/README.md and pkg/timedilation for wiring notes.
package loadpredict
