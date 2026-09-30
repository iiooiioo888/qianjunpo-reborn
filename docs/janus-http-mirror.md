# Janus HTTP tactical mirrors (dev / Cocos browser preview)

Janus `:8090`（宿主機 `make compose-up` 時為 `:18090`) 提供與 gRPC 戰術 API 對應的 JSON 端點，供無 gRPC 的客戶端（例如 Cocos Live preview）使用。

| Method | Path | gRPC equivalent |
|--------|------|-----------------|
| `POST` | `/v1/tactical/connect` | `Connect` |
| `POST` | `/v1/tactical/enter-battle` | `EnterBattle` |
| `GET` | `/v1/tactical/snapshot?battle_id=…` | `GetBattleSnapshot` |
| `POST` | `/v1/tactical/command` | `SubmitTacticalCommand` |

## 必要順序：Connect → EnterBattle → 指令／快照（全 HTTP）

Roma 的戰局存在**記憶體分區**中，只有經 Janus **`EnterBattle`**（內部呼叫 Roma `JoinZone`）才會建立 `battle_id`（例如 `default/0`）。若略過 EnterBattle 直接 `POST /v1/tactical/command`，會得到 HTTP 200 但 **`accepted: false`**，`reject_reason` 通常為 **`roma: battle not found`**。

瀏覽器 Live 預覽建議流程（**不需 grpcurl**）：

1. **`POST /v1/tactical/connect`** — 取得 `session_id`（與正式 gRPC 客戶端一致）
2. **`POST /v1/tactical/enter-battle`** — 建立戰局，回傳 `battle_id`、`initial_state_hash`、`view_snapshot_json`
3. **`POST /v1/tactical/command`** 或 **`GET /v1/tactical/snapshot`**

整合測試見 `services/janus/http_tactical_test.go`：`TestHTTPTacticalEnterBattleThenCommandAccepted`（全 HTTP 路徑）、`TestHTTPTacticalCommandMirror`（gRPC EnterBattle 後 HTTP 指令）、與 `TestHTTPTacticalCommandMirrorWithoutEnterBattle`（未 EnterBattle 時 `battle not found`）。

### curl（`make compose-up`，宿主機埠）

Janus HTTP 在 **`:18090`**。`access_token` 需與 compose 內 Janus 設定一致（開發環境常用佔位 token；見 `.env` / compose）。

```bash
# 1) Connect
SESSION=$(curl -sS -X POST 'http://127.0.0.1:18090/v1/tactical/connect' \
  -H 'Content-Type: application/json' \
  -d '{"access_token":"dev","target_zone":{"zone_id":"default","shard":0}}' \
  | jq -r '.session_id')

# 2) EnterBattle — 在 Roma 建立戰局，取得 battle_id
BATTLE=$(curl -sS -X POST 'http://127.0.0.1:18090/v1/tactical/enter-battle' \
  -H 'Content-Type: application/json' \
  -d "{\"session_id\":\"$SESSION\",\"access_token\":\"dev\",\"target_zone\":{\"zone_id\":\"default\",\"shard\":0}}" \
  | jq -r '.battle_id')

# 3) POST 戰術指令（應 accepted: true）
curl -sS -X POST 'http://127.0.0.1:18090/v1/tactical/command' \
  -H 'Content-Type: application/json' \
  -d "{\"session_id\":\"$SESSION\",\"battle_id\":\"$BATTLE\",\"player_id\":0,\"kind\":1,\"unit_id\":1,\"to_x\":5,\"to_y\":8}"

# 4) GET 顯示用快照
curl -sS "http://127.0.0.1:18090/v1/tactical/snapshot?battle_id=$BATTLE" -D -
```

若跳過 EnterBattle，僅對 `default/0` 發指令，會與測試 `TestHTTPTacticalCommandMirrorWithoutEnterBattle` 相同而遭拒絕。

### grpcurl（選用）

仍可用 gRPC **`:19090`** 做 Connect / EnterBattle，再對 **`:18090`** 發 HTTP 指令／快照；見歷史 PR #57 與 `grpcurl` 文件。

## POST `/v1/tactical/connect`

Request：

```json
{
  "access_token": "dev",
  "client_version": "optional",
  "target_zone": { "zone_id": "default", "shard": 0 }
}
```

Response：

```json
{
  "session_id": "…",
  "server_time": { "wall_unix_ms": 0, "sim_tick": 0 },
  "roma_endpoint": "roma:9092"
}
```

## POST `/v1/tactical/enter-battle`

Request：

```json
{
  "session_id": "…",
  "access_token": "dev",
  "target_zone": { "zone_id": "default", "shard": 0 }
}
```

Response：

```json
{
  "battle_id": "default/0",
  "initial_state_hash": 0,
  "sim_time": { "wall_unix_ms": 0, "sim_tick": 0 },
  "view_snapshot_json": { }
}
```

`view_snapshot_json` 為 Roma 戰術顯示層 JSON（與 `GET /v1/tactical/snapshot` body 同型）。

## POST `/v1/tactical/command`

Request body（snake_case，與 `TacticalCommandClient.ts` 對齊）：

```json
{
  "session_id": "optional",
  "battle_id": "default/0",
  "player_id": 0,
  "kind": 1,
  "unit_id": 1,
  "to_x": 5,
  "to_y": 8
}
```

Response：

```json
{
  "accepted": true,
  "reject_reason": "",
  "lockstep_frame": 0,
  "state_hash": 0
}
```

成功時 `state_hash` 為非零；`lockstep_frame` 為 Roma 戰局目前 frame。

### Environment

| Variable | Default | Meaning |
|----------|---------|---------|
| `JANUS_HTTP_TACTICAL_COMMAND` | enabled (`1`) | Set to `0` / `false` / `disabled` to return HTTP 503 with `accepted: false` instead of forwarding to Roma. |
