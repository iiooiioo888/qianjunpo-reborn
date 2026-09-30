# 戰鬥公式與配置（Phase 2 加深）

## 配置載入

- 預設檔：`config/combat/combat.json`（`combat.DefaultConfig()` 內嵌快照）。
- 對局建立時呼叫 `combat.LoadFile` / `combat.LoadBytes` 或沿用 `DefaultConfig()`，並在 `tactical.NewMatch` / `NewMatchWithConfig` **一次性**寫入 `Match` 的 counter 矩陣與兵種目錄。
- **模擬路徑禁止** `fixed.FromFloat`；克制倍率以 JSON `num`/`den` 整數比載入，經 `fixed.FromInt` + `Div` 轉成 FP64。
- 局中使用中的配置不可熱替換而不破壞確定性；若要換表請在新對局 `NewMatch` 前載入。

## 克制矩陣

`counters.matrix[attacker][defender]` 為攻方兵種對守方兵種的倍率（3×3，順序：步兵 / 弓兵 / 騎兵）。

預設為剪刀石頭布：步兵＞弓兵、弓兵＞騎兵、騎兵＞步兵（強 1.25、弱 0.8、中性 1.0）。

## 最終攻擊與傷害

與白皮書 Phase 2 垂直切片一致（§4.4 細節未入庫時採可測的簡化式）：

1. `FinalATK = BaseATK × counter[attacker.Type][defender.Type]`
2. `Damage = max(0, FinalATK − defender.BaseDEF)`
3. `HP' = HP − Damage`

戰術層攻擊距離：目標格與攻擊者 Chebyshev 距離須滿足 `1 ≤ d ≤ Range`（`Range` 來自兵種目錄）。

## 終局條件（`pkg/tactical`）

| 條件 | `EndReason` | `Winner` |
|------|-------------|----------|
| 僅一方尚有 HP>0 單位 | `EndAnnihilation` | 存活方 player ID |
| 雙方皆無存活單位 | `EndMutualWipe` | `NoWinner`（平手） |
| 達 `maxTurnFrames` 仍未殲滅 | `EndTimeout` | 總 HP 較高者；同 HP 則 `NoWinner` |

佔點勝尚未實作；本切片僅殲滅與超時比 HP。
