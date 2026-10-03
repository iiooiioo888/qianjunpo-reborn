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

**占點（occupy）** 需 `NewMatchWithControlPoints`（Roma Join 預設對局不帶據點；測試會在 store 內替換 match）。P0 出生格 `(2,8)` 若為唯一據點且 `HoldFrames=2`，連續兩步 lockstep（無需指令）即終局：`endReason: "occupy"`、`winner: 0`。

**超時（timeout）** 在雙方每幀 `kind=3` pass 時於 `maxTurnFrames`（64）觸發。HP 較高者勝（`winner: 0|1`）；同 HP 和局為 `winner: 255`、`endReason: "timeout"`。

## 觀戰／自動（spectator / auto）

無玩家手動 `Submit` 時，**仍可直接 `StepLockstep`**，frame 會照常遞增（超時／占點同樣會在規則滿足時終局）。

若需雙方自動出指令（簡單 AI：朝敵方移動、能攻擊則攻擊、否則 pass）：

- `Match.SetAutoCommandMode(AutoCommandBoth)` 後每幀呼叫 `StepWithAuto()`，或
- `RunSpectatorAuto(maxFrames)` 一次跑完。

終局 `winner`／`endReason` token 與手動對局相同（例如殲滅仍為 `"wipeout"`）。

## 測試（驗收）

進行中契約：

```bash
go test ./pkg/tactical/ -run TestViewSnapshotInProgressHasNoWinner
go test ./pkg/tactical/ -run TestInitialViewSnapshotShape
```

殲滅（既有，勿改契約）：

```bash
go test ./pkg/tactical/ -run TestViewSnapshotAnnihilationExportsWinnerAndReason
go test ./internal/roma/ -run TestTacticalViewSnapshotJSONReportsWipeout
go test ./pkg/integration/ -run TestPhase2VictoryViewSnapshotIntegration
go test ./services/janus/ -run TestHTTPTacticalSnapshotReportsWipeoutVictory
```

**占點 smoke**（ViewSnapshot `endReason: "occupy"`、`winner` 為占點方）：

```bash
go test ./pkg/tactical/ -run TestViewSnapshotCaptureExportsOccupy
go test ./internal/roma/ -run TestTacticalViewSnapshotJSONReportsOccupy
go test ./pkg/integration/ -run TestPhase2OccupyViewSnapshotIntegration
go test ./services/janus/ -run TestHTTPTacticalSnapshotReportsOccupyVictory
```

**超時 smoke**（ViewSnapshot `endReason: "timeout"`、`winner` 依 HP 或和局 `255`）：

```bash
go test ./pkg/tactical/ -run 'TestViewSnapshotTimeout|TestTimeout'
go test ./internal/roma/ -run TestTacticalViewSnapshotJSONReportsTimeout
go test ./pkg/integration/ -run TestPhase2TimeoutViewSnapshotIntegration
go test ./services/janus/ -run TestHTTPTacticalSnapshotReportsTimeoutVictory
```

**觀戰／自動**（無手動指令仍推進 frame；簡單 AI 可打到 wipeout）：

```bash
go test ./pkg/tactical/ -run 'TestSpectatorAuto|TestIdleStepLockstep'
```

一次跑齊勝敗快照相關測：

```bash
go test ./pkg/tactical/ -run ViewSnapshot
go test ./internal/roma/ -run 'ViewSnapshotJSONReports'
go test ./pkg/integration/ -run 'ViewSnapshotIntegration|VictoryViewSnapshot'
go test ./services/janus/ -run 'SnapshotReports.*Victory'
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

占點／超時目前以 **上述 `go test` smoke** 驗收（Roma Join 預設無據點；超時需 64 步 pass，不納入 shell smoke）。手動驗證時可改 store 內 match 或本地 `NewMatchWithControlPoints` 後走相同 HTTP 路徑。

整合說明亦見 `docs/combat-formula.md`（占點／勝敗）與 `docs/skill-cast.md`（快照欄位慣例）。
