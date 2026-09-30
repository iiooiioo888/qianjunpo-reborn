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

```bash
make test-edge-infer   # mock HTTP handlers
make test-ai-rag       # pkg/ai + pkg/rag focused tests
```
