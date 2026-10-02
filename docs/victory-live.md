# 勝敗條件 — Live ViewSnapshot

權威邏輯在 `pkg/tactical`（殲滅 `checkVictory`、占點 `tickCapture`、回合上限 `decideTimeout`）。Roma 每步 `StepLockstep` 後，`MatchToViewSnapshot` 序列化進 Janus HTTP／gRPC 的 **`view_snapshot_json`**（與 `/qjp/?live=1` 預覽相同 JSON）。

## ViewSnapshot 欄位（`schemaVersion=1`）

| 欄位 | 進行中 | 終局後 |
|------|--------|--------|
| `winner` | `null` | 玩家 id `0` / `1`，或和局／未決 sentinel `255`（`tactical.NoWinner`） |
| `endReason` | `"none"` | `"wipeout"`（殲滅）、`"occupy"`（占點）、`"timeout"`（超時）、`"mutual_wipe"`（雙方同滅） |

`POST /v1/tactical/step-lockstep` 回應另含 **`finished`**、`winner`（與 proto StepLockstep 一致）；顯示層建議以快照內 `winner`／`endReason` 為準，與 `lastSkillCast` 相同模式。

## Live 如何打到終局

預設對局（`NewMatch`）**不含占點**；在 Live 上最容易驗證的是 **殲滅（wipeout）**：

1. 開啟 `/qjp/?live=1`，連線並進入戰局。
2. 將敵方單位 HP 打到 0（移動＋攻擊或 `kind=5` 技能），再 **step-lockstep** 直到該單位從棋盤清除。
3. 下一幀快照應為 `endReason: "wipeout"`、`winner: 0` 或 `1`。

占點模式需 `NewMatchWithControlPoints`（目前 Roma Join 預設對局不帶據點）；超時則在雙方只 pass 時於 `maxTurnFrames`（64）觸發。

## 測試

```bash
go test ./pkg/tactical/ -run 'ViewSnapshot|Annihilation|Capture'
go test ./internal/roma/ -run TestTacticalViewSnapshotJSONReportsWipeout
go test ./pkg/integration/ -run TestPhase2VictoryViewSnapshot
go test ./services/janus/ -run TestHTTPTacticalSnapshotReportsWipeoutVictory
```

## Compose smoke（HTTP 透傳）

`pkg/tactical` 或 Roma／Janus 映像更新後，請先 **重建** 再跑 smoke（否則容器內仍是舊 snapshot schema）：

```bash
docker compose build janus roma
# 或 make compose-up（會依 Dockerfile 重建）
export ACCESS_TOKEN=…   # 見 docs/janus-http-mirror.md
WIPEOUT_SMOKE=1 ./scripts/janus-http-smoke.sh
```

腳本會：enter-battle 斷言進行中 `winner: null`、`endReason: "none"` → bridge 至敵方相鄰格 → 重複 Strike 至殲滅 → 斷言 step-lockstep `view_snapshot_json` 與 `GET /v1/tactical/snapshot` 皆為 `endReason: "wipeout"`、`winner: 0|1`。

瀏覽器 static-preview：`/qjp/?live=1` 依相同 JSON 欄位顯示終局 overlay；離線契約檢查 `node client/scripts/static-preview-battle-end-smoke.mjs`（見 `client/static-preview/README.md`）。

整合說明亦見 `docs/combat-formula.md`（占點／勝敗）與 `docs/skill-cast.md`（快照欄位慣例）。
