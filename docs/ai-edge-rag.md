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

### `POST /v1/rag-infer` (live battle context → Top-K → infer)

Runs the **real** in-process path: seed corpus → vector Top-K (query = `battle_context` + `order`) → augmented prompt → backend (`mock` or `ollama`). Use this for curl verification of `rag_k` / `rag_hit_ids`.

```bash
make edge-infer   # EDGE_INFER_BACKEND=mock by default
curl -s -X POST localhost:8088/v1/rag-infer \
  -H 'Content-Type: application/json' \
  -d '{
    "persona": "guard",
    "order": "flank east with cavalry",
    "battle_context": {
      "units_summary": "three cavalry on the east wing",
      "terrain_summary": "hills favor archers"
    }
  }' | jq
```

Success JSON (mock backend):

```json
{
  "text": "advance: …",
  "model": "qwen2.5-3b-mock",
  "latency_ms": 1,
  "source": "edge",
  "rag_k": 5,
  "rag_hit_ids": ["tactic-flank", "wei-xiahou", "…"]
}
```

| Field | Meaning |
|-------|---------|
| `rag_k` | Top-K depth used for retrieval (`pkg/rag.DefaultTopK`, usually 5) |
| `rag_hit_ids` | Corpus document ids from Top-K (stable offline ids like `tactic-flank`) |

`order` and legacy `prompt` are accepted; at least one is required. `battle_context.units_summary` / `terrain_summary` are optional but should mirror live unit/terrain summaries from the tactical view.

In-process (`pkg/ai.RAGInferClient.InferWithBattle`) returns the same `rag_k` / `rag_hit_ids` on `InferResult` after retrieval, before edge HTTP.

### `POST /v1/infer` success JSON

Snake_case fields mirror `pkg/ai.InferResult` for observability:

```json
{
  "text": "advance: defend the gate",
  "model": "qwen2.5-3b-mock",
  "latency_ms": 1,
  "source": "edge",
  "rag_k": 5,
  "rag_hit_ids": ["tactic-flank", "wei-cao", "shu-zhang", "tactic-spear", "wu-zhou"]
}
```

| Field | Meaning |
|-------|---------|
| `source` | `edge` on successful model output from this service |
| `rag_k` | Top-K depth the client used when retrieving lore (omitted when plain `InferClient.Infer` with no RAG) |
| `rag_hit_ids` | Retrieved chunk ids in rank order; echoed from the POST body when `RAGInferClient` augmented the prompt |
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

`RAGInferClient` retrieves Top-5 (query = optional `BattleContext` unit/terrain summary + order), builds a prompt prefix (`Battlefield:` + `Knowledge:` from `rag.FormatContext`), then calls edge HTTP with the augmented prompt plus `rag_k` / `rag_hit_ids` on the POST body. Use `InferWithBattle` in-process or `POST /v1/rag-infer` on edge-infer for curl verification. Edge-infer echoes those fields on success JSON so operators can prove Top-K entered the real infer path. No RAG or LLM output enters `pkg/fixed` / combat FP64 math.

On successful retrieval, `RAGInferClient` emits structured `slog` at info level (`msg=ai infer rag augmented`) with `rag_k` and `rag_hit_ids` (optional `persona`). Set `RAGInferClient.LogRAG=false` to silence in tests. Example:

```text
{"time":"…","level":"INFO","msg":"ai infer rag augmented","rag_k":5,"rag_hit_ids":["tactic-flank","wei-cao","shu-zhang","tactic-spear","wu-zhou"],"persona":"cavalry"}
```

`InferResult` also carries `RAGK` and `RAGHitIDs` after `InferWithContext` / `InferWithBattle`, including NPC fallback when retrieval succeeded but edge HTTP failed.

`InferResult` fields:

| Field | When set |
|-------|----------|
| `Source` | `edge` (model) or `npc` (template) |
| `RAGK` / `RAGHitIDs` | When RAG Top-K augmented the prompt (even if edge later fell back to NPC) |
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

`make test-ai-rag` includes `pkg/ai/rag_infer_test.go` (`TestRAGInferTopKEntersPromptAndResult`, `TestRAGInferLogsAugmentation`), which assert Top-K chunk text is in the POST prompt, `rag_hit_ids`/`rag_k` on `InferResult` and edge JSON, and structured RAG slog—not only NPC fallback counters (`npc_fallback_metrics_test.go`).

When edge `/v1/infer` fails, the service returns JSON with HTTP 502. Keys align with in-process NPC fallback (`fallback_reason` / `fallback_detail` use the same strings as `npc_fallback_by_reason`):

```json
{
  "error": "backend unavailable",
  "code": "backend_infer_failed",
  "fallback_reason": "http_non_ok_status",
  "fallback_detail": "502:backend_infer_failed"
}
```

`pkg/ai.InferClient` still maps this into `InferResult` with `source=npc` and the same reason/detail (HTTP status path).

### Live HUD consumers

- **Per-infer line**: read `source`, `fallback_reason`, and `fallback_detail` from successful infer JSON (`POST /v1/infer` HTTP 200) or from `InferResultToHTTP` when a gateway wraps `InferClient` / `RAGInferClient` in-process. On edge model success, `source=edge` and fallback keys are omitted. On NPC template text, `source=npc` with a non-empty `fallback_reason` (`timeout`, `http_non_ok_status`, `no_infer_client`, …).
- **RAG on fallback**: when Top-K ran before edge failed, `rag_k` / `rag_hit_ids` may still be present on the same JSON object (see `TestRAGInferSlowEdgeFallsBackWithinDeadline`).
- **Aggregate counters**: `GET /health` → `npc_fallback_by_reason` (same reason strings as `fallback_reason`).
- **Direct edge failure**: HTTP 502 body includes `fallback_reason` / `fallback_detail` as above; HUD can surface them without guessing from status code alone.

Example HUD-facing JSON after RAG + edge timeout (gateway or in-process serialization):

```json
{
  "text": "Hold the line!",
  "source": "npc",
  "fallback_reason": "timeout",
  "fallback_detail": "context deadline exceeded",
  "rag_k": 5,
  "rag_hit_ids": ["tactic-flank", "wei-cao", "shu-zhang", "tactic-spear", "wu-zhou"]
}
```

### `POST /v1/suggest` (staggered edge vs RAG → one tactical suggestion)

Alternates **edge-only** and **RAG-augmented** infer legs (`suggestion.infer_path` is `edge` or `rag`). Returns the usual infer JSON plus a single **move or skill** hint and a top-level **`command`** object aligned with Janus `POST /v1/tactical/command` (auto-battle pollers can POST `command` as-is). Optional `battle_id` **write-back** stores the latest suggestion for polling.

**Auto-battle consumption** (no guessing):

| Field | Role |
|-------|------|
| `suggestion.kind` | Human/debug: `move` or `skill` |
| `suggestion.unit_id`, `to_x`, `to_y`, `skill_id` | Same semantics as tactical command |
| `suggestion.infer_path` | `edge` or `rag` (observability) |
| `command.kind` | `1` = move, `5` = skill (`KindMove` / `KindSkill`) |
| `command.player_id` | Always `0` for player-0 suggestions |
| `command.battle_id` | Echoes request `battle_id` when set |

When edge infer is down or the model is unavailable, the handler still returns **HTTP 200** with `source: "npc"`, non-empty `text`, plus **`suggestion` and `command`** derived from the NPC template (stub tactical move/skill).

**Clickable test page** (edge-infer only; does not use `static-preview` overlay/tile/battle-end):

```bash
make edge-infer
# open in browser:
#   http://localhost:8088/v1/suggest/demo
```

The page calls the same-origin `POST /v1/suggest` and `GET /v1/suggest/write-back` buttons and renders `suggestion` (move/skill) plus raw JSON. Production HUD can later embed the same contract without this HTML.

```bash
make edge-infer
curl -s -X POST localhost:8088/v1/suggest \
  -H 'Content-Type: application/json' \
  -d '{
    "battle_id": "default/0",
    "persona": "guard",
    "order": "flank east with cavalry",
    "battle_context": {
      "units_summary": "three cavalry on the east wing",
      "terrain_summary": "hills favor archers"
    }
  }' | jq
```

Example success (RAG leg may include `rag_k` / `rag_hit_ids`):

```json
{
  "text": "advance: …",
  "model": "qwen2.5-3b-mock",
  "latency_ms": 1,
  "source": "edge",
  "rag_k": 5,
  "rag_hit_ids": ["tactic-flank", "…"],
  "suggestion": {
    "kind": "move",
    "unit_id": 101,
    "to_x": 8,
    "to_y": 8,
    "infer_path": "rag"
  },
  "command": {
    "battle_id": "default/0",
    "player_id": 0,
    "kind": 1,
    "unit_id": 101,
    "to_x": 8,
    "to_y": 8
  }
}
```

| `suggestion.kind` | `command.kind` |
|-------------------|----------------|
| `move` | `1` (KindMove) |
| `skill` | `5` (KindSkill); `command.skill_id` required (stub strike = `1`) |

Auto-tick loop (mirror): poll suggest, then submit command to Janus:

```bash
# 1) Fetch suggestion + command (edge-infer)
curl -s -X POST localhost:8088/v1/suggest \
  -H 'Content-Type: application/json' \
  -d '{"battle_id":"default/0","persona":"guard","order":"flank east"}' | jq '.command'

# 2) POST the command object to Janus (after EnterBattle)
curl -sS -X POST 'http://127.0.0.1:18090/v1/tactical/command' \
  -H 'Content-Type: application/json' \
  -d @- <<'EOF'
{"battle_id":"default/0","player_id":0,"kind":1,"unit_id":101,"to_x":8,"to_y":8}
EOF
```

Poll write-back after `battle_id` was set on POST (includes `command`):

```bash
curl -s 'localhost:8088/v1/suggest/write-back?battle_id=default/0' | jq
```

```bash
make test-edge-infer   # mock HTTP handlers
make test-ai-rag       # pkg/ai + pkg/rag focused tests
```
