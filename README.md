# 千軍破·重生（qianjunpo-reborn）

《千軍破》復刻專案的 **確定性模擬與戰棋核心**（Phase 1 + Phase 2）。Go 邏輯層以定點數（FP64）、可重播 RNG 與權威棋盤驗證建立可鎖步、可回滾的戰局基礎。

> **Go 版本**：簡報建議 Go 1.24+；目前 CI／開發環境使用 **Go 1.22.2**。模組 `go` 指令設為 `1.22` 以相容現有工具鏈。

## 目錄結構

| 路徑 | 職責 |
|------|------|
| `pkg/fixed` | FP64（32.32）定點數與 `Vector2` |
| `pkg/rng` | Xorshift128+，含 `GetState` / `SetState` / `Clone` |
| `pkg/hash` | FNV-1a 64-bit 狀態哈希（Desync 檢測） |
| `pkg/lockstep` | 鎖步常數、指令延遲 3 幀、樂觀空包 |
| `pkg/sim` | 最小戰局引擎、快照、回滾重模擬 |
| `pkg/board` | 19×19 棋盤、地形、八向鄰格、切比雪夫距離 |
| `pkg/pathfind` | A\*、JPS、並查集連通快檢 |
| `pkg/validate` | 伺服器權威移動驗證 |
| `pkg/cmdqueue` | 熱／溫／冷隊列與 P0/P1/P2 優先級 |
| `pkg/combat` | 兵種克制矩陣與 FP64 攻防 |
| `pkg/replay` | 戰鬥回放（哈希鏈 + gzip） |
| `cmd/demo` | 雙客戶端同種子同輸入哈希對照 |

## 鎖步與時間模型

- **LockstepTurn**：10 fps（100ms／幀）
- **指令延遲**：提交後 **3** 個鎖步幀才執行
- **子幀**：每個鎖步幀內 **6** 個 `GameTurnFrame` 邏輯步
- **樂觀幀**：缺指令時以空 `CommandPacket` 補齊

## 技術約束

- 戰棋邏輯與戰鬥數值使用 **整數格點 + FP64 定點**；邏輯層禁止浮點運算（`FromFloat` 僅供初始化）
- 尋路與棋盤判定為確定性整數演算法
- RNG 與狀態哈希可序列化／比對，支援快照、回滾與回放驗證

## 本機測試與 Demo

```bash
make test          # 等同 go test ./...
make demo          # 雙客戶端確定性演示
go test ./... -v   # 詳細輸出
```

## Docker Compose（dev）

依賴服務（Redis AOF、MySQL 8、可選 etcd）與可建置的 `app` 映像（預設在容器內跑 `make test`）。

```bash
cp .env.example .env    # 設定本地 MYSQL_ROOT_PASSWORD
make compose-up         # 啟動 redis / mysql / etcd（profile dev）
make compose-test       # 建置 app 並在容器內執行測試
make compose-down       # 關閉堆疊
```

從宿主機連線時使用下列埠（避免與本機常見的 Redis／MySQL 衝突）；容器內 `app` 仍透過服務名與預設埠連線（`redis:6379`、`mysql:3306`、`etcd:2379`）：

| 服務 | 宿主機埠 |
|------|----------|
| Redis | `16379` |
| MySQL | `13306` |
| etcd | `12379` |

手動：

```bash
docker compose --profile dev up -d redis mysql
docker compose --profile dev build app
```

## 本階段未包含

時間膨脹、AI、完整 Janus/Roma/Lares 微服務業務、K8s/Agones、Cocos 客戶端等（見技術白皮書後續階段）。

## 授權

待定（專案初始化階段）。
