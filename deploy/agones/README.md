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
