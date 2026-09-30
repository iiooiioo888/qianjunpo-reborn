# 指令佇列（`pkg/cmdqueue`）

Phase 2 加深：熱／溫／冷三層 + P0/P1/P2 同優先級 FIFO。

## 容量

| 層級 | 預設 | 說明 |
|------|------|------|
| 熱 | 1000 | 全優先級合計上限（對齊 `timedilation.HighWatermark`） |
| 溫 | 5000 | 全優先級合計上限 |
| 冷 | TTL 5s | 無固定深度上限；逾時条目在 `Dequeue` / `Len` 時清除 |

`DefaultConfig()` 使用上表；單元測試可縮小 `HotCap` / `WarmCap` / `ColdTTL`。

## 出隊順序

1. 溫度：熱 → 溫 → 冷  
2. 同溫下優先級：P0 → P1 → P2  
3. 同溫同優先級：FIFO  

## 溢位（`TryEnqueue`）

1. 先試熱，滿則溫，再滿則冷（冷僅受 TTL 約束）。  
2. **P2**：熱、溫皆滿時直接丟棄（`Dropped: true`），不寫入冷層。  
3. **P1**：熱／溫滿時仍可進冷層；冷層 TTL 到期則靜默丟棄。  
4. **P0**：不丟棄；熱／溫滿時進冷層。冷層 TTL 到期時觸發 `SetP0StaleHook`（回滾路徑信號，本包不實作 lockstep）。  

顯式指定溫度的 `Enqueue` 不做 spill，僅在該層有容量時寫入（冷層 `Enqueue` 不受熱／溫計數限制）。

## 與 Phase 3 的關係

過載合併見 `pkg/cmdmerge`（熱佇列深度 + `time_flow_rate`）。本包只提供深度計數（`HotLen` 等），不耦合 timedilation。
