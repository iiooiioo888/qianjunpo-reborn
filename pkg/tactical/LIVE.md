# Live 戰術預設（Core 接線）

本文件描述 `pkg/tactical` 匯出的 **Live 占點／超時** 預設，供 `internal/roma/zone.go` Join 改為帶控制點的 match 時直接引用。

## 常數

| 名稱 | 值 | 說明 |
|------|-----|------|
| `MaxTurnFrames` | `64` | 超時判定門檻（`frame >= 64`） |
| `LiveOccupyHoldFrames` | `3` | 占點需連續獨占的 lockstep 回合數 |
| `LiveSpawnP0` | `(2, 8)` | P0 預設出生格（步兵） |
| `LiveSpawnP1` | `(16, 10)` | P1 預設出生格（騎兵） |
| `LiveCenterPass` | `(9, 9)` | 河面關卡／橋心 |

## 控制點點集（可直接貼進 `NewMatchWithControlPoints`）

**占點（建議，P0 出生即站在據點上）：**

```go
tactical.LiveControlPointsOccupy()
// 等同：
[]tactical.ControlPoint{
    {Pos: board.Coord{X: 2, Y: 8}, HoldFrames: 3},
}
```

**占點（備選，橋心 9,9）：**

```go
tactical.LiveControlPointsOccupyCenter()
// 等同：
[]tactical.ControlPoint{
    {Pos: board.Coord{X: 9, Y: 9}, HoldFrames: 3},
}
```

**超時：** 不帶控制點（`NewMatch` / `NewLiveMatchTimeout`），依 `MaxTurnFrames` 結束。

## 建構子

| 函式 | 控制點 | `LiveAutoProfile` |
|------|--------|-------------------|
| `NewLiveMatchOccupy(seed)` | `LiveControlPointsOccupy()` | `LiveAutoOccupy` |
| `NewLiveMatchOccupyCenter(seed)` | `LiveControlPointsOccupyCenter()` | `LiveAutoOccupy` |
| `NewLiveMatchTimeout(seed)` | 無 | `LiveAutoTimeout` |
| `NewLiveMatchWithControlPoints(seed, pts)` | 自訂 | `LiveAutoOccupy` |

## Roma / Janus 接線

1. **Join 建立對局**（示例占點）  
   `m := tactical.NewLiveMatchOccupy(seed)`  
   或  
   `m := tactical.NewMatchWithControlPoints(seed, tactical.LiveControlPointsOccupy()); m.SetLiveAutoProfile(tactical.LiveAutoOccupy)`

2. **全自動**  
   `auto_command_mode = 2`（wire `AutoCommandWireBoth` → `AutoCommandBoth`），每 tick 呼叫 `StepWithAuto()`。

3. **占點勝利條件**  
   - 預設點在 P0 出生格：玩家手動佈陣後，P0 單位留在 `(2,8)` 且格上僅一方存活單位時，全自動會 **Pass 累積獨占**，約 **3 個 lockstep 回合** 內 `endReason: "occupy"`（早於典型 wipeout ~43 幀）。  
   - P1 全自動會朝最近據點移動，但若 P0 已連續獨占滿 `HoldFrames` 即終局。

4. **超時路徑**  
   `m := tactical.NewLiveMatchTimeout(seed)` + `auto_command_mode = 2`：雙方 AI **每幀 Pass**，不互殺，對局存活至 **`frame >= 64`** 後 `endReason: "timeout"`（HP 高者勝，同 HP 和局）。

5. **預設 wipeout 路徑（不變）**  
   `NewMatch(seed)` + `LiveAutoDefault`（未設定 profile）+ 全自動 → 仍為追逐交戰 wipeout。

## 測試

- `TestLiveAutoOccupyBeforeWipeout` — 全自動占點早於 wipeout  
- `TestLiveAutoTimeoutReachesMaxTurnFrames` — 全自動撐滿超時  
- `go test ./pkg/tactical/...`
