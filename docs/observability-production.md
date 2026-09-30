# Production monitoring (P2)

Prometheus metrics live in `pkg/observability/metrics` and are exposed on Janus/Roma/loadpredict HTTP `/metrics`.

## Stable metric names

| Name | Type | Meaning |
|------|------|---------|
| `qjp_online_players` | gauge | Gateway-side online count (Janus stub) |
| `qjp_battle_latency_p99_ms` | gauge | Battle path latency stub (Roma) |
| `qjp_queue_len` | gauge | Pending command queue driving timedilation |
| `qjp_time_flow_rate` | gauge | Sim time rate (`1.0` = normal, from 10000 fixed-point) |
| `qjp_active_rooms` | gauge | In-memory battles on Roma partition |
| `process_cpu_seconds_total` | counter | Go process CPU (stdlib collector) |

## Alert thresholds (skeleton)

File: `deploy/observability/prometheus/alerts/qjp-production.rules.yml`

| Alert | Condition (starting point) |
|-------|----------------------------|
| `QjpHighCPU` | `rate(process_cpu_seconds_total[5m]) > 0.80` for 5m |
| `QjpQueueDepthHigh` | `qjp_queue_len > 500` for 2m |
| `QjpTimeFlowDegraded` | `qjp_time_flow_rate < 0.50` for 1m |
| `QjpActiveRoomsHigh` | `qjp_active_rooms > 1000` for 5m |

Import `deploy/observability/grafana-phase5-dashboard.json` into Grafana with a Prometheus datasource.

## Local / CI checks

```bash
make observability-check      # Go tests: exposition + artifact references
make promtool-check-rules     # promtool or Docker prom/prometheus image
make check-observability    # both
```

No secrets or cluster credentials are required for these targets.
