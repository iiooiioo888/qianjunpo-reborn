# Observability (Phase 5 skeleton)

Prometheus metric names (namespace `qjp`):

| Metric | Type | Source |
|--------|------|--------|
| `qjp_online_players` | gauge | Janus stub |
| `qjp_battle_latency_p99_ms` | gauge | Roma stub |
| `qjp_queue_len` | gauge | timedilation / load sample |
| `qjp_time_flow_rate` | gauge | timedilation (10000=1.0x) |
| `qjp_active_rooms` | gauge | Roma in-memory battles |
| `process_cpu_seconds_total` | counter | Go process CPU (Prometheus collector) |

Alert rules: `prometheus/alerts/qjp-production.rules.yml` — validate with `make promtool-check-rules`.

Import `grafana-phase5-dashboard.json` into Grafana when a Prometheus datasource is available.
CI does **not** start Prometheus or Grafana; use `make check-observability` for syntax/metric checks.

See also `docs/observability-production.md`.

OpenTelemetry: `pkg/observability/trace` — `StartJanusToRomaSpan` for client spans (noop tracer in tests).

Structured logging field conventions are documented in the root README Phase 5 section.
