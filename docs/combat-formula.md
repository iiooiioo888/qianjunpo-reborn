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
2. 傷害（`combat.ResolveDamage`，由 `damage.armor_k` 決定）：
   - `armor_k = 0`（預設）：`Damage = max(0, FinalATK − defender.BaseDEF)`
   - `armor_k > 0`：`Damage = max(0, FinalATK × K / (K + defender.BaseDEF))`
3. 命中（可選，`damage.hit_rate`；預設未配置則必中、不消耗 RNG）：
   - `hit_rate` 為整數比 `num/den`：戰術層以 `roll % den < num` 判定命中
   - 未命中時 `Damage = 0`（不再進行暴擊判定）
4. 暴擊（可選，`damage.crit_rate` / `damage.crit_mul`；預設未配置則關閉、不消耗 RNG）：
   - `crit_rate` 為整數比 `num/den`：戰術層以 `roll % den < num` 判定暴擊
   - 暴擊時 `Damage = Damage × crit_mul`（整數比，FP64）
5. `HP' = HP − Damage`

戰術層攻擊距離（`combat.InAttackRange` / `tactical.validateAttack`）：

- 目標格與攻擊者 Chebyshev 距離須滿足 `1 ≤ d ≤ Range`（`Range` 來自兵種目錄）。
- **遠程**（目錄 `range > 1`，如弓兵）：另須 `combat.AttackLineClear`——沿 8 向直線步進的射線上每一格地形可通行，且除攻擊者／目標單位外不得有第三方單位占位（與 `pathfind.ExpandSegment` 同 trace）。
- **近戰**（`range ≤ 1`）：僅 Chebyshev 距離，不檢查射線。

## 終局條件（`pkg/tactical`）

| 條件 | `EndReason` | `Winner` |
|------|-------------|----------|
| 僅一方尚有 HP>0 單位 | `EndAnnihilation` | 存活方 player ID |
| 雙方皆無存活單位 | `EndMutualWipe` | `NoWinner`（平手） |
| 達 `maxTurnFrames` 仍未殲滅 | `EndTimeout` | 總 HP 較高者；同 HP 則 `NoWinner` |

佔點勝尚未實作；本切片僅殲滅與超時比 HP。

## 範圍傷害選格與套用（AoE stub）

群傷／範圍技能完整系統尚未入庫；目前提供 Chebyshev 選格，並在戰術層對目標**逐一**走與單體攻擊相同的傷害管線（`ResolveDamage` → 可選 `hit_rate` → 可選 `crit`；不改公式本體）。

- `combat.CollectAoECells(center, radius)`：回傳 Chebyshev 距離 `≤ radius` 的合法格（預設 stub 常數 `DefaultAoERadius = 1`，含中心格與八鄰），按 `Y` 再 `X` 排序以保確定性。
- `combat.CollectAoETargets(board, center, radius, excludeID)`：在上述格子上收集單位 ID（升序），可排除施放者；僅選格，不套用傷害。
- `combat.ResolveStrikeDamage`：純函式版單次結算（供測試與文件對照）；戰術層 `applyStrike` 與單體 `KindAttack` 共用同一 RNG 消耗順序。
- 戰術層 `(*tactical.Match).ApplyAoEStrike(attackerID, center, radius)`：對 `CollectAoETargets` 結果中的**敵方**單位升序逐個 `applyStrike`；友軍在範圍內略過。
- 指令 `KindAoE`（`Command.To` = 範圍中心）：提交時驗證中心在盤內（`AOE_OUT_OF_BOUNDS`）且至少一名敵人在 splash（`AOE_NO_TARGETS`）；**不**改 `InAttackRange`／`AttackLineClear`／`CollectAoECells` 行為；單體 `validateAttack` 路徑不變。
