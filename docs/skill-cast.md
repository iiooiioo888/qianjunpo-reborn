# 技能施放（KindSkill）— Phase 2 最小切片

## 如何 cast（權威路徑）

在已建立的 `tactical.Match` 上提交 `KindSkill` 指令，經 lockstep 延遲後執行（與移動／攻擊相同）：

```go
err := m.Submit(tactical.Command{
    PlayerID: 0,
    Kind:     tactical.KindSkill,
    UnitID:   tactical.UnitIDPlayer0,
    To:       board.Coord{X: tx, Y: ty}, // 目標格
    SkillID:  combat.SkillStubStrike,    // 或 combat.SkillStubSplash
})
// 推進 lockstep.CommandDelayFrames+1 次 StepLockstep 後傷害落地
```

- **卡池**：僅 `combat.DefaultCardPool()` 內技能可施放（`SkillStubStrike`、`SkillStubSplash`）。
- **行為**：`SkillStubStrike` 沿用單體攻擊校驗／`applyStrike`；`SkillStubSplash` 沿用 `KindAoE` 校驗／`ApplyAoEStrike`。
- **Replay**：`tactical.Encode` payload v2 在座標後附 `skill_id`（見 `pkg/tactical/command.go`）。

## Snapshot 可見什麼

`MatchToViewSnapshot` / Roma `GetTacticalViewSnapshot` 回傳的 JSON（`schemaVersion=1`）在技能**執行後**包含：

| 欄位 | 說明 |
|------|------|
| `units[].hp` | 受擊方 HP 下降（傷害效果） |
| `lastSkillCast`（可選） | 最近一次成功解析的 `KindSkill`：`skillId`、`casterUnitId`、`targetX`/`targetY`、`lockstepFrame` |

`lastSkillCast` 僅供顯示層標記施放，**不**參與 `StateHash`。

整合測：`TestSubmitKindSkillStrikeViewSnapshotShowsCastAndDamage`（`pkg/tactical/skill_test.go`）、`TestPhase2SkillCastViewSnapshotIntegration`（`pkg/integration/skill_cast_snapshot_test.go`）。

## Roma / Janus / proto 缺口（本次未改 wire）

| 層 | 現狀 |
|----|------|
| `pkg/tactical` | ✅ `KindSkill` + `SkillID` 已打通 Submit → lockstep → 傷害 |
| `proto/gateway` `SubmitTacticalCommandRequest` | ❌ 無 `skill_id` 欄位；`kind=5` 無法帶技能 id |
| `proto/roma` `TacticalCommand` | ❌ 註解仍為 move/attack/pass；無 `skill_id` |
| `internal/roma/protoToCommand` | ❌ 僅接受 move/attack/pass；`KindSkill` 會被拒 |
| Janus HTTP `POST /v1/tactical/command` | ❌ JSON body 無 `skill_id`（見 `services/janus/http_tactical.go`） |

Live／HTTP 施放技能需核心協調新增 `skill_id` 並擴充 Roma 轉換；本 PR 以 **程式內 `Match.Submit`** 與整合測驗收。詳見 `docs/combat-formula.md` 技能 stub 小節。
