# 千軍破·重生（qianjunpo-reborn）

《千軍破》復刻專案的 **確定性模擬核心**（Phase 1）。目標是在 Go 邏輯層內以定點數與可重播 RNG 建立可鎖步、可回滾的戰局狀態機基礎。

> **Go 版本**：簡報建議 Go 1.24+；目前 CI／開發環境使用 **Go 1.22.2**（見 `go version`）。模組 `go` 指令設為 `1.22` 以相容現有工具鏈。

## 目錄結構

| 路徑 | 職責 |
|------|------|
| `pkg/fixed` | FP64（32.32）定點數與 `Vector2` |
| `pkg/rng` | Xorshift128+，含 `GetState` / `SetState` / `Clone` |
| `pkg/hash` | FNV-1a 64-bit 狀態哈希（Desync 檢測） |
| `pkg/lockstep` | 鎖步常數、指令延遲 3 幀、樂觀空包 |
| `pkg/sim` | 最小戰局引擎、快照、回滾重模擬 |
| `cmd/demo` | 雙客戶端同種子同輸入哈希對照 |

## 鎖步與時間模型

- **LockstepTurn**：10 fps（100ms／幀）
- **指令延遲**：提交後 **3** 個鎖步幀才執行
- **子幀**：每個鎖步幀內 **6** 個 `GameTurnFrame` 邏輯步
- **樂觀幀**：缺指令時以空 `CommandPacket` 補齊

## 技術約束

- 邏輯層 **禁止浮點**；`FromFloat` 僅供初始化配置
- 乘法使用 128-bit 中間結果；除零回傳 `MaxValue`
- RNG 與狀態哈希可序列化／比對，支援快照與回滾

## 執行測試與 Demo

```bash
make test          # 等同 go test ./...
make demo          # 雙客戶端確定性演示
go test ./... -v   # 詳細輸出
```

## 本階段未包含

19×19 棋盤完整玩法、A*、微服務、Cocos 客戶端、時間膨脹、AI 等（見技術白皮書後續階段）。

## 授權

待定（專案初始化階段）。
