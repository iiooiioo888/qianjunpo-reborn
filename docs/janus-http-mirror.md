# Janus HTTP tactical mirrors (dev / Cocos browser preview)

Janus `:8090` (宿主機 `make compose-up` 時為 `:18090`) 提供與 gRPC 戰術 API 對應的 JSON 端點，供無 gRPC 的客戶端（例如 Cocos Live preview）使用。

| Method | Path | gRPC equivalent |
|--------|------|-----------------|
| `GET` | `/v1/tactical/snapshot?battle_id=…` | `GetBattleSnapshot` |
| `POST` | `/v1/tactical/command` | `SubmitTacticalCommand` |

## 必要順序：先 EnterBattle，再 HTTP 指令／快照

Roma 的戰局存在**記憶體分區**中，只有經 Janus **`EnterBattle`**（內部呼叫 Roma `JoinZone`）才會建立 `battle_id`（例如 `default/0`）。若略過這一步直接 `POST /v1/tactical/command`，會得到 HTTP 200 但 **`accepted: false`**，`reject_reason` 通常為 **`roma: battle not found`**。

建議流程：

1. **gRPC `Connect`**（取得 `session_id`，可選但與正式客戶端一致）
2. **gRPC `EnterBattle`**（建立戰局並回傳 `battle_id`、`initial_state_hash`、初始 `view_snapshot_json`）
3. **`POST /v1/tactical/command`** 或 **`GET /v1/tactical/snapshot`**

整合測試見 `services/janus/http_tactical_test.go`：`TestHTTPTacticalCommandMirror`（EnterBattle 後 `accepted: true`）與 `TestHTTPTacticalCommandMirrorWithoutEnterBattle`（未 EnterBattle 時 `battle not found`）。

### grpcurl + curl（`make compose-up`，宿主機埠）

Janus gRPC 在 **`:19090`**，HTTP 在 **`:18090`**。以下 `access_token` 需與 compose 內 Janus 設定一致（開發環境常用佔位 token；見 `.env` / compose）。

```bash
# 1) Connect（可選 session，建議與客戶端一致）
SESSION=$(grpcurl -plaintext -d '{
  "access_token": "dev",
  "target_zone": {"zone_id": "default", "shard": 0}
}' 127.0.0.1:19090 qianjunpo.gateway.v1.JanusGateway/Connect | jq -r '.sessionId')

# 2) EnterBattle — 在 Roma 建立戰局，取得 battle_id
BATTLE=$(grpcurl -plaintext -d "{
  \"session_id\": \"$SESSION\",
  \"access_token\": \"dev\",
  \"target_zone\": {\"zone_id\": \"default\", \"shard\": 0}
}" 127.0.0.1:19090 qianjunpo.gateway.v1.JanusGateway/EnterBattle | jq -r '.battleId')

# 3) POST 戰術指令（應 accepted: true）
curl -sS -X POST 'http://127.0.0.1:18090/v1/tactical/command' \
  -H 'Content-Type: application/json' \
  -d "{\"session_id\":\"$SESSION\",\"battle_id\":\"$BATTLE\",\"player_id\":0,\"kind\":1,\"unit_id\":1,\"to_x\":5,\"to_y\":8}"

# 4) GET 顯示用快照
curl -sS "http://127.0.0.1:18090/v1/tactical/snapshot?battle_id=$BATTLE" -D -
```

若跳過步驟 2，僅對 `default/0` 發指令，會與測試 `TestHTTPTacticalCommandMirrorWithoutEnterBattle` 相同而遭拒絕。

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
