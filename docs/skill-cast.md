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

## Live / HTTP 施放（Janus → Roma）

`kind` = **`5`**（`tactical.KindSkill`）時 body 必帶 **`skill_id`**（stub：`1` = Strike、`2` = Splash，見 `pkg/combat/skills.go`）。移動／攻擊／pass 可省略 `skill_id`。

```bash
curl -sS -X POST 'http://127.0.0.1:18090/v1/tactical/command' \
  -H 'Content-Type: application/json' \
  -d '{"battle_id":"default/0","player_id":0,"kind":5,"unit_id":101,"to_x":3,"to_y":8,"skill_id":1}'
```

步進後 `GET /v1/tactical/snapshot` 或 `POST /v1/tactical/step-lockstep` 的 `view_snapshot_json` 會保留 **`lastSkillCast`** 與 **`units[].hp`**（與 Roma `GetTacticalViewSnapshot` 相同 JSON，HTTP 層不做欄位白名單過濾）。詳見 `docs/janus-http-mirror.md`。

**Compose smoke**：`SKILL_ID=1 ./scripts/janus-http-smoke.sh` 會先 bridge 移至敵方相鄰格（預設對局 **101→(15,10)** 再對 **(16,10)** 施放 Strike），再 step-lockstep 斷言 `lastSkillCast` / HP。

| 層 | 現狀 |
|----|------|
| `pkg/tactical` | ✅ Submit → lockstep → 傷害 + snapshot |
| `proto/gateway` / `proto/roma` | ✅ `skill_id` 欄位 |
| `internal/roma/protoToCommand` | ✅ `KindSkill` + `SkillID` |
| Janus HTTP `POST /v1/tactical/command` | ✅ JSON `skill_id` |
