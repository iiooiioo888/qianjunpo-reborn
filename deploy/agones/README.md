# Agones skeleton (Phase 4 — not applied in CI)

These manifests document how Roma game servers would run as an Agones **Fleet** with
**Buffer** autoscaler and **Counter**-based matchmaking hooks. Replace image/registry
before applying to a live cluster.

## Files

| File | Purpose |
|------|---------|
| `fleet.yaml` | Agones Fleet wrapping Roma container |
| `buffer-autoscaler.yaml` | Buffer capacity autoscaler |
| `counter.yaml` | Counter CR skeleton for queue depth |

Apply manually when an Agones-enabled cluster is available:

```bash
kubectl apply -f deploy/agones/
```

## Phase 5 — predict → scale → rate recover

1. **Predict**: `pkg/loadpredict` / `services/loadpredict` stub forecasts 30s-ahead load; fires when composite load ≥ 90%.
2. **Scale**: `PreScaleSignal` targets Agones **buffer-autoscaler.yaml** / Fleet capacity 7–15 minutes ahead of the spike (skeleton only in CI).
3. **Recover**: After queue drains, Roma `pkg/timedilation` raises `time_flow_rate` (see `RecoverTimeFlow` hint on signals).

Wire load samples from Roma command queues into `loadsample.Sample` for the control loop.

## Pre-prod room lifecycle (Go)

Roma integrates **Allocate → Ready → Shutdown** via `pkg/agones` (mock in CI, sidecar REST on Fleet pods). See **[docs/agones-room-lifecycle.md](../../docs/agones-room-lifecycle.md)** for env vars, CLI (`cmd/agones-room`), HTTP routes (`/v1/agones/*`), and acceptance steps.

Suggested Fleet env for game server pods:

```yaml
env:
  - name: ROMA_AGONES_BACKEND
    value: sidecar
  - name: ROMA_AGONES_ALLOCATOR
    value: http
  - name: ROMA_AGONES_ALLOCATION_URL
    value: https://allocation.example/agones/v1/gameserverallocation
```
