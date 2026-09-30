# Roma / 國戰 Agones pre-prod room lifecycle

Pre-production wiring for **Allocate → Ready → Shutdown** on Roma game servers. CI uses the **mock** backend (same pattern as `EDGE_INFER_BACKEND=mock`); Dev clusters use the Agones **sidecar HTTP** gateway and optional **GameServerAllocation** HTTP endpoint.

## Environment variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `ROMA_AGONES_BACKEND` | `mock` | GameServer lifecycle: `mock` or `sidecar` |
| `ROMA_AGONES_ALLOCATOR` | `mock` | Control-plane allocate: `mock` or `http` |
| `ROMA_AGONES_SIDECAR_URL` | `http://127.0.0.1:9358` | Sidecar REST base (real pods) |
| `ROMA_AGONES_ALLOCATION_URL` | _(empty)_ | POST target for GameServerAllocation (required when allocator=`http`) |
| `ROMA_AGONES_ALLOCATION_TOKEN` | _(empty)_ | Bearer token for allocation API |
| `ROMA_AGONES_NAMESPACE` | `default` | K8s namespace |
| `ROMA_AGONES_FLEET` | `roma-fleet` | Fleet label selector (`deploy/agones/fleet.yaml`) |
| `ROMA_AGONES_*_TIMEOUT` | allocate 30s / ready 15s / shutdown 30s | Per-step deadlines |
| `ROMA_AGONES_HEALTH_INTERVAL` | `2s` | Sidecar health ping interval in Roma |
| `ROMA_AGONES_GAMESERVER_PORT` | `9092` | Mock allocate response gRPC port |

Disable Agones hooks entirely: `ROMA_AGONES_DISABLE=1` (Roma skips lifecycle; HTTP admin routes omitted).

## Local (mock)

```bash
export ROMA_AGONES_BACKEND=mock
export ROMA_AGONES_ALLOCATOR=mock

# Full CLI path
go run ./cmd/agones-room allocate -zone dev -shard 0
go run ./cmd/agones-room ready
go run ./cmd/agones-room status
go run ./cmd/agones-room shutdown

# Or run Roma (auto Ready after gRPC listen + mock health loop)
ROMA_GRPC_ADDR=:19092 ROMA_HTTP_ADDR=:18092 ETCD_DISABLE=1 \
  go run ./services/roma
curl -s http://127.0.0.1:18092/v1/agones/status | jq .
curl -s -X POST http://127.0.0.1:18092/v1/agones/allocate \
  -H 'Content-Type: application/json' \
  -d '{"zone_id":"dev","shard":0}' | jq .
```

## Dev cluster (real sidecar)

1. Apply Fleet manifests: `kubectl apply -f deploy/agones/`
2. Set Roma container env: `ROMA_AGONES_BACKEND=sidecar`, `ROMA_AGONES_ALLOCATOR=http`
3. Point `ROMA_AGONES_ALLOCATION_URL` at your allocation service (Agones allocator or in-cluster proxy).
4. Roma calls sidecar `POST /ready` after listeners start and `POST /shutdown` on SIGTERM.

Sidecar REST reference: [Agones Client SDKs — REST](https://agones.dev/site/docs/guides/client-sdks/rest/).

## Acceptance checklist

1. `make test-agones-room` passes (mock Allocate → Ready → Shutdown).
2. `make test` stays green.
3. `go run ./cmd/agones-room allocate && ready && shutdown` exits 0 with mock env.
4. Roma `/v1/agones/status` returns JSON with `phase=ready` after startup (mock or sidecar).
5. Forced failures: set mock delays / use short `ROMA_AGONES_READY_TIMEOUT` — logs contain `agones/lifecycle:` lines and `last_error` in status JSON.
6. Prometheus: `qjp_agones_lifecycle_operations_total{operation="ready",result="ok"}` increments on successful Ready.

## Make targets

```bash
make test-agones-room   # pkg/agones unit tests
make agones-room-demo   # mock allocate → ready → shutdown CLI
```

See also `deploy/agones/README.md` for Fleet / buffer manifests.
