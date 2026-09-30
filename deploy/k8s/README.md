# Roma on Kubernetes (Phase 4 skeleton)

Roma runs as a **StatefulSet** with a **headless Service** for stable per-pod DNS
(`roma-0.roma-headless`, …). Battle authoritative state remains **in pod memory**;
Redis is not used for HP/position/buff in this design.

## Resource notes (request = limit)

For predictable latency on deterministic simulation partitions, set **requests equal to limits**
(no burstable CPU). Example in `roma-statefulset.yaml`:

- CPU: `request: 2000m`, `limit: 2000m`
- Memory: `request: 4Gi`, `limit: 4Gi`

Adjust per zone load; these values are placeholders for Agones/K8s sizing discussions.

## Apply (not run in CI)

```bash
kubectl apply -f deploy/k8s/
```
