# Agones pre-prod scale-out (Agones 1.54.0)

Manifests for Roma as an Agones **Fleet** with **Buffer** or **Counter** FleetAutoscaler policies and a **players** counter (CountsAndLists Beta). Replace image/registry before applying to a live cluster.

**Room lifecycle** (Allocate → Ready → Shutdown): [docs/agones-room-lifecycle.md](../../docs/agones-room-lifecycle.md)  
**Buffer / Counter / Allocate wiring**: [docs/agones-scale-out.md](../../docs/agones-scale-out.md)

## Files

| File | Purpose |
|------|---------|
| `fleet.yaml` | Fleet + `counters.players` + Roma sidecar env |
| `buffer-autoscaler.yaml` | **Default** pre-prod autoscaler — Buffer of Ready GameServers |
| `fleetautoscaler-counter.yaml` | Optional — scale on free `players` capacity (do not apply with buffer autoscaler) |
| `gameserverallocation-players.yaml` | Example allocation with counter filter + increment |
| `counter.yaml` | Index / comments for the `players` counter key |

## Apply order (dev cluster)

```bash
# 1. Fleet (declares counters.players)
kubectl apply -f deploy/agones/fleet.yaml

# 2. Pick one autoscaler policy
kubectl apply -f deploy/agones/buffer-autoscaler.yaml
# OR (not both):
# kubectl apply -f deploy/agones/fleetautoscaler-counter.yaml

# 3. Point Roma allocator at your allocation service; use gameserverallocation-players.yaml as body reference
```

Quick apply of the default bundle (Fleet + Buffer only):

```bash
kubectl apply -f deploy/agones/fleet.yaml -f deploy/agones/buffer-autoscaler.yaml
```

## Agones 1.54.0 notes

- **CountsAndLists** is Beta and enabled by default in 1.54; counters must be declared on the Fleet template before use.
- **Buffer** autoscaler maintains spare **Ready** replicas; **Counter** autoscaler maintains spare aggregate capacity on key `players`.
- Prometheus: `agones_fleet_counters` exposes fleet-wide counter totals when the counter policy or controller metrics are enabled.

## Phase 5 — predict → scale → rate recover

1. **Predict**: `pkg/loadpredict` forecasts load; `PreScaleSignal.TargetBuffer` hints desired buffer size.
2. **Scale**: Tune `buffer-autoscaler.yaml` `bufferSize` / `maxReplicas` (or counter `bufferSize` / `maxCapacity`) from those hints — skeleton only in CI.
3. **Recover**: Roma `pkg/timedilation` may raise `time_flow_rate` after queue drains (`RecoverTimeFlow` on signals).

Wire load samples from Roma command queues into `loadsample.Sample` for the control loop.

## Local / CI verification (no cluster)

```bash
make validate-agones-manifests   # structural checks on YAML under deploy/agones/
make test-agones-room            # pkg/agones unit tests (mock lifecycle)
make check-agones                # both of the above
```

Suggested Fleet env for game server pods (also embedded in `fleet.yaml`):

```yaml
env:
  - name: ROMA_AGONES_BACKEND
    value: sidecar
  - name: ROMA_AGONES_ALLOCATOR
    value: http
  - name: ROMA_AGONES_ALLOCATION_URL
    value: https://allocation.example/agones/v1/gameserverallocation
```
