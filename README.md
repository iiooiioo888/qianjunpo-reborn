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
| `pkg/replay` | 戰鬥回放（哈希鏈 + gzip；v2 含每幀 `time_flow_rate`） |
| `pkg/timedilation` | 時間膨脹控制環、`time_flow_rate`、分層時鐘、追趕上限 1.5x |
| `pkg/loadsample` | CPU／隊列／成長率／P99 採樣 + `Predictor`（LSTM 可插拔介面） |
| `pkg/degrade` | L0–L5 降級狀態機與有序恢復 |
| `pkg/cmdmerge` | 過載指令合併（move/build、P2 500ms 批次） |
| `pkg/ai` | 戰略層（~5s mock）+ 戰術層 → `lockstep.CommandPacket` |
| `pkg/rag` | RAG Top-K 介面 + 記憶體假向量庫（Top-5 延遲目標見套件註解） |
| `services/edge-infer` | 邊緣推理 HTTP 骨架（health + Qwen2.5-3B mock） |
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

### Phase 3：時間膨脹與 AI 邊界

- **模擬幀率**仍固定 **10 fps**；`time_flow_rate` 以萬分比（10000=1.0x）調節區域／角色牆鐘對齊，**不進入** `pkg/fixed`／`pkg/combat` 的 FP64 戰鬥運算
- 分層時鐘：`real` → `region` → `actor`；`battle` 幀計數只隨鎖步 +1
- 回放 **v2** 在每幀旁記錄 `TimeFlowRates`；v1 仍可讀
- 過載時佇列高水位 1000、低水位 200，漸進降速／較慢恢復（單元測試以 `ManualClock` 驗證 ~4.5s／~11.6s 量級）

## 本機測試與 Demo

```bash
make test          # 等同 go test ./...
make demo          # 雙客戶端確定性演示
make edge-infer    # 啟動 :8088 邊緣推理 mock（/health、/v1/infer、/v1/load）
go test ./pkg/timedilation -v
go test ./pkg/ai ./pkg/integration -v
go test ./services/edge-infer -v
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

完整 Janus/Roma/Lares 微服務業務、真實 Milvus/Qdrant、下載多 GB 模型、K8s/Agones、Cocos 客戶端 UI 等（見技術白皮書後續階段）。

## 授權

待定（專案初始化階段）。
