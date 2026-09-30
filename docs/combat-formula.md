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
- **近戰**（`range ≤ 1`）：僅 Chebyshev 距離，不檢查射線（`combat.RangedAttackNeedsLine` 為 false；與 #69／#72 遠程阻擋盤面對照時，同一直線延伸段上的第三方占位或不可通行地形**不**拒絕近戰 `KindAttack`，整合測見 `TestSubmitKindAttackMeleeLoSNotCheckedUnitOnLineSubmitDamagesEnemy`、`TestSubmitKindAttackMeleeLoSNotCheckedTerrainOnLineSubmitDamagesEnemy`）。
- **遠程單體 `KindAttack` 與 LOS**：射線暢通時 `validateAttack` 通過並在 lockstep 執行後對目標套傷；被第三方單位或不可通行地形擋住則 `validate.CodeBlocked`（暢通見 `TestSubmitKindAttackRangedLoSClearDamagesEnemy`；單位占位阻擋見 `TestSubmitKindAttackLoSBlockedRejected`；射線上不可通行地形見 `TestSubmitKindAttackLoSTerrainBlockedRejected`；盤面與 #64 AoE 對稱）。

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
- 指令 `KindAoE`（`Command.To` = 範圍中心）：提交與 `ApplyAoEStrike` 前，對中心格套用與單體攻擊相同的 `combat.InAttackRange`／`combat.AttackLineClear`（遠程才檢 LOS；中心格占位單位 ID 作為射線忽略的 target）；失敗碼 `AOE_OUT_OF_RANGE`／`AOE_LOS_BLOCKED`。另驗證中心在盤內（`AOE_OUT_OF_BOUNDS`）且至少一名敵人在 splash（`AOE_NO_TARGETS`）。**不**改 `InAttackRange`／`AttackLineClear`／`CollectAoECells` 本體；單體 `validateAttack` 路徑不變。
- **中心格占位**：近戰／遠程皆不可選施放者所在格（`InAttackRange` 要求 `d ≥ 1`）。中心可落在**友軍**格或空格；只要射程／LOS 合法且 splash 內有敵人即可施放，傷害僅對敵方單位升序 `applyStrike`，友軍在範圍內不扣血（整合測見 `TestApplyAoEStrikeFriendlyCenterSkipsAllies`、`TestSubmitKindAoEFriendlyCenterLockstepIntegration`）。
- **遠程 AoE 與 LOS**：射線暢通時 `KindAoE` 與單體相同地通過 `AttackLineClear` 並在 lockstep 執行後對中心／splash 內敵人套傷；被第三方單位或不可通行地形擋住則 `AOE_LOS_BLOCKED`（暢通見 `TestSubmitKindAoERangedLoSClearDamagesEnemy`；單位占位阻擋見 `TestSubmitKindAoELoSBlockedRejected`；射線上不可通行地形見 `TestSubmitKindAoELoSTerrainBlockedRejected`；盤面與 #69 KindAttack 地形測對稱）。
- **近戰 AoE 與 LOS**：與單體近戰相同，不跑 `AttackLineClear`；對照遠程阻擋盤面見 `TestSubmitKindAoEMeleeLoSNotCheckedUnitOnLineSubmitDamagesEnemy`、`TestSubmitKindAoEMeleeLoSNotCheckedTerrainOnLineSubmitDamagesEnemy`。
