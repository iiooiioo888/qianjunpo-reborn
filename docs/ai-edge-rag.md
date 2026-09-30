# Edge inference & RAG (Phase 3 deepen)

## Edge service (`services/edge-infer`)

| Env | Default | Meaning |
|-----|---------|---------|
| `EDGE_INFER_BACKEND` | `mock` | `mock` (CI) or `ollama` |
| `EDGE_INFER_ADDR` | `:8088` | HTTP listen address |
| `EDGE_INFER_OLLAMA_URL` | `http://127.0.0.1:11434` | Ollama / OpenAI-compatible base |
| `EDGE_INFER_OLLAMA_MODEL` | `qwen2.5:3b` | Model tag (e.g. Q4_K_M quant via Ollama) |
| `EDGE_INFER_HEALTH_PROBE_TIMEOUT` | `2s` | Max wait for Ollama `/api/tags` on `/health` (infer still uses long client) |

Client (`pkg/ai.InferClient`):

| Env | Default | Meaning |
|-----|---------|---------|
| `AI_INFER_REQUEST_TIMEOUT` | `1s` | Whole infer attempt before NPC fallback |
| `AI_INFER_HTTP_TIMEOUT` | `800ms` | HTTP client limit per POST `/v1/infer` |

### Local Ollama

```bash
ollama pull qwen2.5:3b
EDGE_INFER_BACKEND=ollama make edge-infer
curl -s localhost:8088/health | jq
curl -s -X POST localhost:8088/v1/infer -d '{"prompt":"defend the gate"}' | jq
```

### `POST /v1/infer` success JSON

Snake_case fields mirror `pkg/ai.InferResult` for observability:

```json
{
  "text": "advance: defend the gate",
  "model": "qwen2.5-3b-mock",
  "latency_ms": 1,
  "source": "edge"
}
```

| Field | Meaning |
|-------|---------|
| `source` | `edge` on successful model output from this service |
| `fallback_reason` | Omitted on edge success; NPC reasons appear when a proxy returns `source=npc` (client-side fallback does not use this HTTP shape) |
| `fallback_detail` | Omitted unless `fallback_reason` is set |

`pkg/ai.InferClient` decodes the success body into `InferResult` (`Source`, `FallbackReason`, `FallbackDetail`). When edge HTTP fails, the client still returns NPC template text with `source=npc` and a populated `fallback_reason` / `fallback_detail` (not an HTTP 200 from edge-infer).

Example client-side NPC fallback after HTTP 502:

```json
{
  "text": "Eyes on the pass—they're moving east.",
  "source": "npc",
  "fallback_reason": "http_non_ok_status",
  "fallback_detail": "502:backend_infer_failed"
}
```

(The above is the in-process `InferResult` shape from `InferClient.Infer` / `RAGInferClient.InferWithContext`, not the edge service response.)

When `backend=ollama` and Ollama is down, `/health` returns HTTP 200 with `status=degraded` and `ready=false` (soft fail). `pkg/ai.InferClient` still falls back to NPC templates within **1s** if `/v1/infer` errors.

### Measuring latency (aspirational targets)

- **TTFT**: time from POST `/v1/infer` to first response byte (use `latency_ms` in JSON or `curl -w '%{time_starttransfer}'`).
- **RAG retrieve**: `go test ./pkg/rag -bench=. -benchtime=3s` or wrap `Store.TopKText` with your metrics sink.

## RAG (`pkg/rag`)

- `ThreeKingdomsSeed()` + `SeedStore()` load factions / generals / tactics into `InMemoryStore`.
- `HashBagEmbedder` provides deterministic offline vectors; swap `Embedder` for a remote API later.
- **Out of scope**: Milvus / Qdrant clusters in this PR.

## Game client wiring (`pkg/ai`)

`RAGInferClient` retrieves Top-5, builds a prompt prefix, then calls edge HTTP. No RAG or LLM output enters `pkg/fixed` / combat FP64 math.

`InferResult` fields:

| Field | When set |
|-------|----------|
| `Source` | `edge` (model) or `npc` (template) |
| `FallbackReason` | Non-empty when `Source=npc` (`timeout`, `http_non_ok_status`, `no_infer_client`, …) |
| `FallbackDetail` | Optional (`502:backend_infer_failed`, transport error text) |

**Registry-aligned persona keys** (match `client` `CharacterCardSpriteRegistry` / `UnitSpriteRegistry`):

| Key | Client registry |
|-----|-----------------|
| `char_caocao` | v04 曹操卡 |
| `char_zhangfei` | v04 張飛卡 |
| `char_wu_placeholder` | v04 吳占位卡 |
| `infantry` | 步兵單位貼圖 |
| `cavalry` | 騎兵單位貼圖 |

Legacy aliases still resolve (e.g. `cao_cao` → `char_caocao`, `zhang_fei` / `zhangfei` → `char_zhangfei`). Generic tactical roles `guard`, `scout`, `strategist`, `merchant` remain for non-card NPCs. Older general ids (`guan_yu`, `liu_bei`, …) keep their lines but are not v04 card keys. Spaces/hyphens normalize to underscores. Unknown personas get `"For the realm!"`.

On NPC fallback, `InferClient` and `RAGInferClient` emit structured `slog` at info level (`msg=ai infer npc fallback`) with `source=npc`, `fallback_reason`, and `fallback_detail` (optional `persona`). Example log line:

```text
{"time":"…","level":"INFO","msg":"ai infer npc fallback","source":"npc","fallback_reason":"http_non_ok_status","fallback_detail":"502:backend_infer_failed","persona":"scout"}
```

Set `InferClient.LogFallback=false` to silence in tests.

### NPC fallback counters (`pkg/ai`)

Each in-process NPC fallback (`source=npc`) increments an atomic counter keyed by `fallback_reason` (`timeout`, `http_non_ok_status`, `no_infer_client`, …). Read totals with `ai.NPCFallbackCounts()` or, when this process also runs `services/edge-infer`, from `GET /health`:

```json
{
  "status": "ok",
  "model": "qwen2.5-3b-mock",
  "backend": "mock",
  "ready": true,
  "npc_fallback_by_reason": {
    "http_transport_error": 2,
    "http_non_ok_status": 1
  }
}
```

Counters reflect fallbacks in **this OS process** (game gateway, tests, or a co-located edge-infer binary that links `pkg/ai`). A standalone edge-infer with no `InferClient` calls usually shows an empty map. After NPC fallbacks in that process, verify with `curl -s localhost:8088/health | jq '.npc_fallback_by_reason'` (e.g. `"no_infer_client": 1` when `RAGInferClient.Infer` is nil).

When edge `/v1/infer` fails, the service returns JSON `{"error","code":"backend_infer_failed"}` with HTTP 502; the client maps status + code into `FallbackDetail`.

```bash
make test-edge-infer   # mock HTTP handlers
make test-ai-rag       # pkg/ai + pkg/rag focused tests
```
