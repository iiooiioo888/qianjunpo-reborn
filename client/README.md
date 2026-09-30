# 千軍破·重生 — Cocos Creator 客戶端 Shell

最小 **Cocos Creator 3.8.x** 工程：本地顯示 19×19 戰術棋盤、單位占位、**time_flow_rate** HUD 掛點。權威邏輯仍在 Go `pkg/tactical` / Roma；本工程僅 **顯示層**（白皮書邏輯／顯示分離）。

## 需求

- [Cocos Creator 3.8.4](https://www.cocos.com/creator-download)（3.8.x 皆可；`package.json` 標記 3.8.4）
- Mock 模式無需啟動 Janus／Roma；Live 模式需本機 `make compose-up` 或手動跑 Janus + Roma

## 開啟專案

1. 啟動 Cocos Dashboard → **新增／打開** → 選本目錄 `client/`（含 `package.json` 的資料夾）。
2. 首次開啟若提示升級引擎版本，選 **3.8.x** 並允許編譯 TypeScript。
3. 在 **資源管理器** 雙擊 `assets/scenes/TacticalBoard.scene`。
4. 點 **預覽**（瀏覽器或模擬器）：應看到 19×19 地格（河／關隘橋）、兩枚占位單位、左上角 `time_flow_rate`／幀號 stub。

若場景腳本綁定遺失：在 `Canvas/TacticalRoot` 上 **添加组件 → 自定義脚本 → TacticalBootstrap**，`Snapshot Resource` 填 `data/tactical/demo_initial`。

## Mock vs Live Janus

| 模式 | `TacticalBootstrap` | 資料來源 |
|------|---------------------|----------|
| **Mock（預設）** | `useLiveJanus = false` | `resources/data/tactical/demo_initial.json`（`make client-snapshot`） |
| **Live（開發）** | `useLiveJanus = true`，`liveBattleId = default/0` | Janus HTTP 開發鏡像 → Roma 權威 `Match` |

### Mock 資料

```bash
# 在倉庫根目錄
make client-snapshot
```

JSON schema 與 `pkg/tactical.ViewSnapshot` 一致：`schemaVersion=1`，`boardSize=19`，`timeFlowRateParts` 為萬分比（10000=1.0x）。

### Live 快照（gRPC 權威路徑）

1. 啟動服務（Compose 預設 Janus gRPC `:9090`、HTTP `:8090`，Roma `:9092`）：

   ```bash
   make compose-up
   ```

2. **EnterBattle** 與 **GetBattleSnapshot**（`proto/gateway/v1/gateway.proto`）會回傳 `view_snapshot_json`（與 mock 同 schema）。`StepTacticalLockstep` 回應亦帶當幀快照。

   ```bash
   # 範例：取得 default 分片戰局初始快照（需 grpcurl）
   grpcurl -plaintext -d '{"access_token":"dev","target_zone":{"zone_id":"default","shard":0}}' \
     localhost:9090 qianjunpo.gateway.v1.JanusGateway/EnterBattle
   grpcurl -plaintext -d '{"battle_id":"default/0"}' \
     localhost:9090 qianjunpo.gateway.v1.JanusGateway/GetBattleSnapshot
   ```

3. **Cocos 瀏覽器預覽**（無 gRPC 插件時）：Janus HTTP 開發端點（與 `GetBattleSnapshot` 同源 JSON，非第二套棋盤）：

   ```bash
   curl -s 'http://127.0.0.1:8090/v1/tactical/snapshot?battle_id=default/0'
   ```

   在場景中勾選 `useLiveJanus`，`liveBattleId` 填 `default/0`（或 EnterBattle 回傳的 `battle_id`）。預設 host／port 見 `JanusGatewayStub.ts` 的 `DEFAULT_NETWORK_STUB`。

### 仍為 stub／後續

- Janus **TCP** 二進位 framing（`:7000`）仍為 skeleton；正式客戶端應走 gRPC／未來 WebSocket。
- Cocos 內尚未內建 gRPC 客戶端；Live 預覽用 HTTP 鏡像，CI／工具用 `grpcurl` 或 `make janus-roma-test`。
- `TimeFlowHudStub` 尚未接心跳／replay `TimeFlowRates` 串流。

## 目錄與分層

| 路徑 | 職責 |
|------|------|
| `assets/scripts/logic/` | 純資料：`ViewSnapshot` 解析（無權威、無 Cocos 依賴） |
| `assets/scripts/display/` | 棋盤／單位占位／色板／HUD stub |
| `assets/scripts/network/` | `fetchLiveViewSnapshot`、Janus 設定 stub |
| `assets/scripts/app/` | `TacticalBootstrap` 場景入口 |
| `assets/resources/data/tactical/` | Mock 戰局 JSON |
| `assets/resources/textures/2d/` | 正式 2D 像素貼圖匯入位（見 [`../art/2d/README.md`](../art/2d/README.md)） |

## 美術（卡通像素風）

- 正式資產：`art/2d/**` 過審後拷貝至 `assets/resources/textures/2d/`；匯入規範見 [`art/2d/README.md`](../art/2d/README.md)。
- 目前棋盤為 **Graphics 色塊 + 占位幾何**；接圖後改 `TacticalBoardView` / `UnitPlaceholderView` 即可，logic 層不變。

## 建置產物

`library/`、`temp/`、`build/` 已在 `.gitignore`；請勿提交。

## 相關後端

- 戰術核心：`pkg/tactical`
- CLI：`cmd/match`、`make match-play`、`make client-snapshot`
- 閘道契約：`proto/gateway/v1/gateway.proto`、`proto/roma/v1/roma.proto`
- 整合測試：`make janus-roma-test`
