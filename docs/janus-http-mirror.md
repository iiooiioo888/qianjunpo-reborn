# Janus HTTP tactical mirrors (dev / Cocos browser preview)

Janus `:8090`（宿主機 `make compose-up` 時為 `:18090`) 提供與 gRPC 戰術 API 對應的 JSON 端點，供無 gRPC 的客戶端（例如 Cocos Live preview）使用。

| Method | Path | gRPC equivalent |
|--------|------|-----------------|
| `POST` | `/v1/lares/login` | Lares gRPC `Login`（Janus HTTP 鏡像，供 Live 預覽） |
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

### Lares Login HTTP 鏡像（同域 Live 預覽，非字面 `dev`）

瀏覽器 **`/qjp/?live=1`** 可同域呼叫 **`POST /v1/lares/login`**（nginx 反代至 Janus HTTP base，例如 `http://127.0.0.1:18090/v1/lares/login`），取得與 **LaresAuth** 相同 HMAC 格式的短期 **`access_token`**。任意非空帳密即可（smoke 腳本與預覽預設 **`smoke` / `smoke`**）；**沒有**接受字面字串 `dev` 作為 token 的路徑。

| 環境變數 | Compose 典型值 | 說明 |
|----------|----------------|------|
| `JANUS_HTTP_DEV_MINT` | `1`（可關 `0`） | 關閉時 **`POST /v1/lares/login`** 回 **404** |
| `LARES_TOKEN_SECRET` | 與 Lares/Janus 共用 | 簽章密鑰（`internal/lares.TokenIssuer`） |

Request：

```json
{
  "username": "smoke",
  "password": "smoke"
}
```

Response（UI 至少讀 **`access_token`**；其餘字段與 Lares `Login` 對齊，snake_case）：

```json
{
  "access_token": "<Lares-signed access token>",
  "refresh_token": "…",
  "access_expires_unix": 1700000900,
  "refresh_expires_unix": 1700604800,
  "player_id": 123456789,
  "account_id": "smoke"
}
```

Access TTL 目前 **15 分鐘**。生產環境請保持 **`JANUS_HTTP_DEV_MINT=0`**。

```bash
export ACCESS_TOKEN=$(curl -sS -X POST 'http://127.0.0.1:18090/v1/lares/login' \
  -H 'Content-Type: application/json' \
  -d '{"username":"smoke","password":"smoke"}' \
  | jq -r '.access_token')
```

### 取得可用的 `ACCESS_TOKEN`（curl / smoke）

1. **建議**：在本機 shell 匯出 **`ACCESS_TOKEN`**（或 `export ACCESS_TOKEN=…`），所有 curl 與腳本共用，**不要把真 token 寫進 repo**。
2. **Compose / 同域預覽**：`POST /v1/lares/login`（見上一節；需 `JANUS_HTTP_DEV_MINT=1`）。
3. **簽發 — Lares gRPC `Login`（任意非空帳密即可；player_id 由帳號衍生）：

   ```bash
   export ACCESS_TOKEN=$(grpcurl -plaintext -d '{"username":"smoke","password":"smoke"}' \
     127.0.0.1:19091 qianjunpo.lares.v1.LaresAuth/Login \
     | jq -r '.accessToken // .access_token')
   ```

   `LARES_TOKEN_SECRET` 須與 compose 內 Janus/Lares 一致（預設見上表；可自 `.env` 覆寫）。

4. **簽章格式（除錯用）**：Lares 使用 **stub token**（非 production JWT），形狀為  
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

**KindSkill（可選）**：設定 `SKILL_ID=1` 或 `MOVE_KIND=5` 時走技能 E2E：從快照解析敵方座標（預設對局騎兵 **201 @ (16,10)**），若我方單位不在相鄰格則重複 **bridge move**（`kind=1` → `POST /v1/tactical/step-lockstep`，預設 `LOCKSTEP_STEPS=4`）至目標旁格（預設 **(15,10)**），再 `kind=5` + `skill_id` 施放並步進，斷言 `lastSkillCast` 與敵方 HP 下降。可覆寫 `SKILL_TO_X`/`SKILL_TO_Y`、`BRIDGE_TO_X`/`BRIDGE_TO_Y`。預設移動 smoke（未設 `SKILL_ID` 且 `MOVE_KIND=1`）行為不變。

```bash
SKILL_ID=1 ./scripts/janus-http-smoke.sh
# 等同 MOVE_KIND=5 SKILL_ID=1；見 docs/skill-cast.md
```

**Wipeout 勝敗（`WIPEOUT_SMOKE=1`）**：在 KindSkill bridge 路徑上重複 Strike 至終局，斷言 `view_snapshot_json`／`GET /v1/tactical/snapshot` 的 `winner`（`0`/`1`）與 `endReason: "wipeout"`；進入戰局時斷言 `winner: null`、`endReason: "none"`。需 **Compose 重建** `janus`／`roma` 後執行（見 `docs/victory-live.md`）。

```bash
docker compose build janus roma
WIPEOUT_SMOKE=1 ./scripts/janus-http-smoke.sh
```

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

`view_snapshot_json` 為 Roma 戰術顯示層 JSON（與 `GET /v1/tactical/snapshot` body 同型）；`units[].id` 為下指令時應使用的 `unit_id`（預設對局常見 **101** / **201**）。技能執行後 JSON 含 **`units[].hp`** 與可選 **`lastSkillCast`**（#89）；HTTP 直接透傳 Roma bytes，不剝欄位。

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

技能施放（`kind` = **5**，`KindSkill`）需附加 **`skill_id`**（例如 `1` = `SkillStubStrike`）：

```json
{
  "session_id": "optional",
  "battle_id": "default/0",
  "player_id": 0,
  "kind": 5,
  "unit_id": 101,
  "to_x": 3,
  "to_y": 8,
  "skill_id": 1
}
```

`kind` 為 move/attack/pass 時可省略 `skill_id`。`kind=5` 且缺 `skill_id` 時 Roma 拒絕（`reject_reason` 含 `skill_id`）。

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
  "steps": 1,
  "auto_command": false,
  "auto_command_mode": 0
}
```

`steps` 省略或 `0` 時視為 **1**。

**觀戰／自動對戰（#96 playable）**

| 欄位 | 意義 |
|------|------|
| `auto_command` | 本次請求每步使用 Roma 簡單 AI 補指令（`StepWithAuto`），與手動 `command` 可並用 |
| `auto_command_mode` | `0` 不變；`1` 關閉；`2` 雙方自動（持久）；`3` 僅補未提交方 |

無玩家輸入時仍可 `POST step-lockstep`（`steps:1`），**frame 會遞增**（超時／占點照常）。推薦觀戰迴圈：一次 `auto_command_mode: 2` 啟用持久自動，之後每 tick 僅 `steps: 1`；或 Infer 路徑 `POST /v1/suggest` → `POST /v1/tactical/command` → `POST step-lockstep`（可不帶 `auto_command`）。

Response：

```json
{
  "lockstep_frame": 1,
  "state_hash": 123,
  "finished": false,
  "winner": 0,
  "auto_command_mode": 1,
  "view_snapshot_json": { "autoCommandMode": "both", "endReason": "none" }
}
```

`view_snapshot_json.autoCommandMode` 為 HUD 用 token：`off` \| `both` \| `missing`。`auto_command_mode` 為 Roma 持久策略 wire（`0=off, 1=both, 2=missing`）。

與 gRPC `StepTacticalLockstep` 對齊；`view_snapshot_json` 為步進後顯示層快照（避免多一次 snapshot RPC）。

**curl smoke（idle frame + 觀戰自動）**

```bash
# enter-battle 後，無 command 仍步進
curl -sS -X POST "$JANUS/v1/tactical/step-lockstep" \
  -H 'Content-Type: application/json' \
  -d '{"battle_id":"default/0","steps":1}' | jq '.lockstep_frame,.finished'

# 啟用持久雙方自動後步進至終局（本地 mock Roma 或 Compose）
curl -sS -X POST "$JANUS/v1/tactical/step-lockstep" \
  -H 'Content-Type: application/json' \
  -d '{"battle_id":"default/0","auto_command_mode":2,"steps":0}'
```

### Environment

| Variable | Default | Meaning |
|----------|---------|---------|
| `JANUS_HTTP_TACTICAL_COMMAND` | enabled (`1`) | Set to `0` / `false` / `disabled` to return HTTP 503 with `accepted: false` instead of forwarding to Roma. |
| `JANUS_HTTP_DEV_MINT` | `0`（Compose 預設 `1`） | Enable `POST /v1/lares/login` HTTP mirror on Janus. |
