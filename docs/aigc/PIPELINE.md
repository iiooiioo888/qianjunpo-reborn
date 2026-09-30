# AIGC pipeline stub (Phase 5)

Whitepaper v6.0 **生產調優** — art/content generation without GPU in CI.

## Placeholders

| Endpoint | Service | Purpose |
|----------|---------|---------|
| `GET /health` | `services/aigc-worker` | Liveness |
| `POST /v1/comfy/workflow` | aigc-worker | ComfyUI workflow submit (stub JSON) |
| `POST /v1/ipadapter/style` | aigc-worker | IP-Adapter reference id (no weights in repo) |

## Makefile

```bash
make aigc-worker    # :8096 stub HTTP
make aigc-stub-check # curl /health only (CI-safe)
```

## Asset import path

1. Export PNG/WebP from ComfyUI or manual art to `assets/import/<category>/` (gitignored large blobs).
2. Register manifest entry in `assets/import/manifest.json` (stub path — create locally).
3. Senate ops approves via `pkg/contentops` queue before hot-reload into client bundles.

**Do not** commit multi-GB checkpoints. CI never downloads models.

## Production notes

- Run ComfyUI on GPU nodes separately; aigc-worker forwards jobs over internal network.
- IP-Adapter weights live on object storage; worker stores opaque `ref_id` only.
