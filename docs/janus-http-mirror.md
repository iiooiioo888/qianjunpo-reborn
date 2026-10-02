# Janus HTTP tactical mirrors (dev / Cocos browser preview)

Janus `:8090`（宿主機 `make compose-up` 時為 `:18090`) 提供與 gRPC 戰術 API 對應的 JSON 端點，供無 gRPC 的客戶端（例如 Cocos Live preview）使用。

| Method | Path | gRPC equivalent |
|--------|------|-----------------|
| `POST` | `/v1/auth/dev-mint` | （無；Lares `TokenIssuer` 本機簽發，供 Live 預覽） |
| `POST` | `/v1/tactical/connect` | `Connect` |
| `POST` | `/v1/tactical/enter-battle` | `EnterBattle` |
| `GET` | `/v1/tactical/snapshot?battle_id=…` | `GetBattleSnapshot` |
| `POST` | `/v1/tactical/command` | `SubmitTacticalCommand` |
| `POST` | `/v1/tactical/step-lockstep` | `StepTacticalLockstep` |

## 認證（Docker Compose：LaresAuth，非 StaticAuth）

`make compose-up` 啟動的 **Janus** 使用 **LaresAuth** 驗證 `access_token`（見 `services/janus/main.go`）：

| 環境變數 | Compose 典型值 | 說明 |
|----------|----------------|------|
| `JANUS_LARES_GRPC_ADDR` | `lares:9091`（宿主機 gRPC **`:19091`**） | Janus 連 Lares Validate |
| `LARES_TOKEN_SECRET` | `${LARES_TOKEN_SECRET:-phase4-dev-secret}` | Janus 與 Lares **共用**簽章密鑰（`internal/lares.TokenIssuer`） |

Janus 會以相同 secret 做本機 HMAC 校驗，並可透過 Lares gRPC `Validate` 備援。**沒有**「任意字串 `dev` 即通過」的 StaticAuth 路徑；對 Compose 送 `"access_token":"dev"` 通常會在 Connect 階段失敗（例如 HTTP **502**，body 含 `janus: unauthorized`）。

### Dev-stable mint（同域 Live 預覽，非字面 `dev`）

瀏覽器 **`/qjp/?live=1`** 可同域呼叫 **`POST /v1/auth/dev-mint`**（nginx 反代至 Janus HTTP），取得與 **LaresAuth** 相同 HMAC 格式的短期 **`access_token`**。預設帳號 **`qjp-live-preview`**（`player_id` 穩定）；**沒有**接受字面字串 `dev` 作為 token 的路徑。

| 環境變數 | Compose 典型值 | 說明 |
|----------|----------------|------|
| `JANUS_HTTP_DEV_MINT` | `1`（可關 `0`） | 關閉時端點回 **404** |
| `LARES_TOKEN_SECRET` | 與 Lares/Janus 共用 | 簽章密鑰（`internal/lares.TokenIssuer`） |
| `JANUS_DEV_MINT_ACCOUNT` | （省略＝`qjp-live-preview`） | 覆寫預設 dev-stable 帳號 |

Request（body 可省略或 `{}`）：

```json
{ "account": "optional-override" }
```

Response：

```json
{
  "access_token": "<Lares-signed access token>",
  "access_expires_unix": 1700000900,
  "player_id": 123456789,
  "account_id": "qjp-live-preview"
}
```

TTL 與 Lares `Login` access 相同（目前 **15 分鐘**）。生產環境請保持 **`JANUS_HTTP_DEV_MINT=0`**。

```bash
export ACCESS_TOKEN=$(curl -sS -X POST 'http://127.0.0.1:18090/v1/auth/dev-mint' \
  -H 'Content-Type: application/json' -d '{}' \
  | jq -r '.access_token')
```

### 取得可用的 `ACCESS_TOKEN`（curl / smoke）

1. **建議**：在本機 shell 匯出 **`ACCESS_TOKEN`**（或 `export ACCESS_TOKEN=…`），所有 curl 與腳本共用，**不要把真 token 寫進 repo**。
2. **Compose / 同域預覽**：`POST /v1/auth/dev-mint`（見上一節；需 `JANUS_HTTP_DEV_MINT=1`）。
3. **簽發 — Lares gRPC `Login`（任意非空帳密即可；player_id 由帳號衍生）：

   ```bash
   export ACCESS_TOKEN=$(grpcurl -plaintext -d '{"username":"smoke","password":"smoke"}' \
     127.0.0.1:19091 qianjunpo.lares.v1.LaresAuth/Login \
     | jq -r '.accessToken // .access_token')
   ```

   `LARES_TOKEN_SECRET` 須與 compose 內 Janus/Lares 一致（預設見上表；可自 `.env` 覆寫）。

3. **簽章格式（除錯用）**：Lares 使用 **stub token**（非 production JWT），形狀為  
   `base64url(payload).base64url(hmac-sha256(secret, payload))`，其中  
   `payload` 明文為 `access|<player_id>|<account>|<exp_unix>`（refresh 則為 `refresh|…`）。  
   實作見 `internal/lares/tokens.go`；手動拼 token 時 secret 必須與 **`LARES_TOKEN_SECRET`** 相同。

## 必要順序：Connect → EnterBattle → 指令／快照（全 HTTP）

Roma 的戰局存在**記憶體分區**中，只有經 Janus **`EnterBattle`**（內部呼叫 Roma `JoinZone`）才會建立 `battle_id`（例如 `default/0`）。若略過 EnterBattle 直接 `POST /v1/tactical/command`，會得到 HTTP 200 但 **`accepted: false`**，`reject_reason` 通常為 **`roma: battle not found`**。

瀏覽器 Live 預覽建議流程（**不需 grpcurl** 做戰術步驟，但仍需有效 Lares token）：

1. **`POST /v1/tactical/connect`** — 取得 `session_id`（與正式 gRPC 客戶端一致）
2. **`POST /v1/tactical/enter-battle`** — 建立戰局，回傳 `battle_id`、`initial_state_hash`、`view_snapshot_json`
3. **`POST /v1/tactical/command`** 或 **`GET /v1/tactical/snapshot`**

整合測試見 `services/janus/http_tactical_test.go`：`TestHTTPTacticalEnterBattleThenCommandAccepted`（全 HTTP 路徑）、`TestHTTPTacticalCommandMirror`（gRPC EnterBattle 後 HTTP 指令）、與 `TestHTTPTacticalCommandMirrorWithoutEnterBattle`（未 EnterBattle 時 `battle not found`）。

預設演練戰的單位 id（`pkg/tactical/unit.go`）：**玩家 0 → `unit_id` 101**，**玩家 1 → 201**。curl 範例請使用快照中真實 id，勿用 `unit_id: 1`。

### curl（`make compose-up`，宿主機埠）

先設定 `ACCESS_TOKEN`（見上一節）。Janus HTTP 在 **`:18090`**。

```bash
# 1) Connect
SESSION=$(curl -sS -X POST 'http://127.0.0.1:18090/v1/tactical/connect' \
  -H 'Content-Type: application/json' \
  -d "{\"access_token\":\"$ACCESS_TOKEN\",\"target_zone\":{\"zone_id\":\"default\",\"shard\":0}}" \
  | jq -r '.session_id')

# 2) EnterBattle — 在 Roma 建立戰局，取得 battle_id
BATTLE=$(curl -sS -X POST 'http://127.0.0.1:18090/v1/tactical/enter-battle' \
  -H 'Content-Type: application/json' \
  -d "{\"session_id\":\"$SESSION\",\"access_token\":\"$ACCESS_TOKEN\",\"target_zone\":{\"zone_id\":\"default\",\"shard\":0}}" \
  | jq -r '.battle_id')

# 3) POST 戰術指令（應 accepted: true；unit_id 101 = owner 0 預設單位）
curl -sS -X POST 'http://127.0.0.1:18090/v1/tactical/command' \
  -H 'Content-Type: application/json' \
  -d "{\"session_id\":\"$SESSION\",\"battle_id\":\"$BATTLE\",\"player_id\":0,\"kind\":1,\"unit_id\":101,\"to_x\":5,\"to_y\":8}"

# 4) GET 顯示用快照
curl -sS "http://127.0.0.1:18090/v1/tactical/snapshot?battle_id=$BATTLE" -D -
```

若跳過 EnterBattle，僅對 `default/0` 發指令，會與測試 `TestHTTPTacticalCommandMirrorWithoutEnterBattle` 相同而遭拒絕。

### 一鍵 smoke（`scripts/janus-http-smoke.sh`）

```bash
export ACCESS_TOKEN=…   # Lares Login 取得
./scripts/janus-http-smoke.sh
# 或：JANUS_HTTP_BASE=http://127.0.0.1:18090 ACCESS_TOKEN=… ./scripts/janus-http-smoke.sh
```

腳本會 connect → enter-battle → command；`unit_id` 優先從 `view_snapshot_json.units` 選取目前 `PLAYER_ID`（預設 0）所屬單位，否則退回 **101**。成功時 exit 0 且 `accepted: true`。

### grpcurl（選用）

仍可用 gRPC **`:19090`** 做 Connect / EnterBattle，再對 **`:18090`** 發 HTTP 指令／快照；token 同樣須為 Lares 簽發。見歷史 PR #57 與 `pkg/integration/compose_janus_roma_test.go`。

## POST `/v1/tactical/connect`

Request：

```json
{
  "access_token": "<Lares access token>",
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
  "access_token": "<Lares access token>",
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

`view_snapshot_json` 為 Roma 戰術顯示層 JSON（與 `GET /v1/tactical/snapshot` body 同型）；`units[].id` 為下指令時應使用的 `unit_id`（預設對局常見 **101** / **201**）。

## POST `/v1/tactical/command`

Request body（snake_case，與 `TacticalCommandClient.ts` 對齊）：

```json
{
  "session_id": "optional",
  "battle_id": "default/0",
  "player_id": 0,
  "kind": 1,
  "unit_id": 101,
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

指令被接受後，單位位置要等 lockstep **步進**才會反映在 `view_snapshot_json`（指令有 `CommandDelayFrames` 延遲）。Live 預覽在 `POST /v1/tactical/command` 成功後應 **`POST /v1/tactical/step-lockstep`**（通常每幀 `steps: 1`），再 `GET /v1/tactical/snapshot` 或直接使用 step 回傳的 `view_snapshot_json`。

## POST `/v1/tactical/step-lockstep`

Request：

```json
{
  "session_id": "optional",
  "battle_id": "default/0",
  "steps": 1
}
```

`steps` 省略或 `0` 時視為 **1**。

Response：

```json
{
  "lockstep_frame": 1,
  "state_hash": 123,
  "finished": false,
  "winner": 0,
  "view_snapshot_json": { }
}
```

與 gRPC `StepTacticalLockstep` 對齊；`view_snapshot_json` 為步進後顯示層快照（避免多一次 snapshot RPC）。

### Environment

| Variable | Default | Meaning |
|----------|---------|---------|
| `JANUS_HTTP_TACTICAL_COMMAND` | enabled (`1`) | Set to `0` / `false` / `disabled` to return HTTP 503 with `accepted: false` instead of forwarding to Roma. |
| `JANUS_HTTP_DEV_MINT` | `0`（Compose 預設 `1`） | Enable `POST /v1/auth/dev-mint` for dev-stable Lares tokens. |
| `JANUS_DEV_MINT_ACCOUNT` | `qjp-live-preview` | Default account for dev-mint when body omits `account`. |
