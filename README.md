# 千軍破·重生（qianjunpo-reborn）

《千軍破》復刻專案的 **確定性模擬與戰棋核心**（Phase 1–3）與 **Phase 4 微服務邊界骨架**（白皮書 v6.0 — 國戰與部署）。Go 邏輯層以定點數（FP64）、可重播 RNG 與權威棋盤驗證建立可鎖步、可回滾的戰局基礎；Phase 4 新增 Janus／Roma／Lares 等服務契約與部署腳手架（無需 CI 內真實 K8s/Agones）。

> **Go 版本**：簡報建議 Go 1.24+；目前 CI／開發環境使用 **Go 1.22.2**。模組 `go` 指令設為 `1.22` 以相容現有工具鏈。

## Phase 4 架構概覽

```
                    ┌─────────────┐
  Client (TCP) ───► │   Janus     │  I/O 閘道：連線、限流占位、etcd 發現占位
                    │  (gateway)  │
                    └──────┬──────┘
                           │ gRPC
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
   ┌──────────┐     ┌──────────┐      ┌──────────┐
   │  Lares   │     │   Roma   │      │  Senate  │  維運 / GM（可選）
   │  (auth)  │     │ (battle) │      │  (ops)   │
   └──────────┘     └──────────┘      └──────────┘
         │                 │
         │           權威狀態僅記憶體
         ▼                 ▼
      MySQL            pkg/sim 等
      Redis*           確定性 CPU 核心
      etcd

* Redis 僅路由／快取；禁止快取 HP／座標／Buff 權威值（見 services/roma/README.md）

   ChatServer（可選）— 頻道訊息 gRPC/HTTP 骨架
```

| 服務 | 路徑 | 邊界 |
|------|------|------|
| **Janus** | `services/janus` | 客戶端 I/O；TCP 橋 + gRPC `JanusGateway` |
| **Roma** | `services/roma` | CPU 確定性權威區服；記憶體戰局 |
| **Lares** | `services/lares` | Access/Refresh 雙令牌 |
| **Senate** | `services/senate` | 維運 gRPC/HTTP |
| **ChatServer** | `services/chatserver` | 聊天 gRPC/HTTP |
| **edge-infer** | `services/edge-infer` | Phase 3 邊緣推理 mock |

契約：`proto/` → `gen/go/`（`make proto` 可選再生；CI 使用已提交生成碼）。

部署：`deploy/k8s/`（Roma StatefulSet + Headless Service）、`deploy/agones/`（Fleet / Buffer / Counter 骨架）、`deploy/docker/Dockerfile.service`（多階段 `CGO_ENABLED=0`）。

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
| `pkg/timesync` | Phase 4：Wall/Sim 雙時間戳、跨區映射、跨區凍結 |
| `pkg/anticheat` | Phase 4：特徵抽取 + 模型推理介面（假資料 E2E） |
| `pkg/loadsample` | CPU／隊列／成長率／P99 採樣 + `Predictor`（LSTM 可插拔介面） |
| `pkg/degrade` | L0–L5 降級狀態機與有序恢復 |
| `pkg/cmdmerge` | 過載指令合併（move/build、P2 500ms 批次） |
| `pkg/ai` | 戰略層（~5s mock）+ 戰術層 → `lockstep.CommandPacket` |
| `pkg/rag` | RAG Top-K 介面 + 記憶體假向量庫（Top-5 延遲目標見套件註解） |
| `proto/` / `gen/go/` | gRPC/Protobuf 契約與生成 Go 存根 |
| `services/*` | 微服務可執行骨架 |
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
make proto         # 可選：需本機 protoc；否則使用已提交的 gen/go
make build-services
make edge-infer    # 啟動 :8088 邊緣推理 mock（/health、/v1/infer、/v1/load）
go test ./pkg/timesync ./pkg/anticheat ./internal/lares -v
go test ./pkg/timedilation -v
go test ./pkg/ai ./pkg/integration -v
go test ./services/edge-infer -v
```

## Docker Compose（dev）

依賴服務（Redis AOF、MySQL 8、etcd）與 Phase 4 微服務 **janus / roma / lares**（`--profile dev`）。可選 **senate / chatserver**（額外 `--profile ops`）。

```bash
cp .env.example .env    # 設定 MYSQL_ROOT_PASSWORD、LARES_TOKEN_SECRET
make compose-up         # redis / mysql / etcd / lares / roma / janus
make compose-up-ops     # 另啟 senate + chatserver
make compose-test       # 建置 app 並在容器內執行測試
make compose-down
```

從宿主機連線時使用下列埠（避免與本機常見服務衝突）：

| 服務 | 宿主機埠（示例） |
|------|------------------|
| Redis | `16379` |
| MySQL | `13306` |
| etcd | `12379` |
| Janus HTTP / gRPC / TCP | `18090` / `19090` / `17000` |
| Lares HTTP / gRPC | `18091` / `19091` |
| Roma HTTP / gRPC | `18092` / `19092` |
| Senate HTTP / gRPC | `18093` / `19093`（ops） |
| Chat HTTP / gRPC | `18094` / `19094`（ops） |

容器內服務間仍使用預設埠（如 `redis:6379`、`roma:9092`）。

## 本階段未包含

真實 Agones 叢集 apply、生產 TLS、Janus 10k 連線效能、下載多 GB 模型、Cocos 客戶端 UI 等。

## 授權

待定（專案初始化階段）。
