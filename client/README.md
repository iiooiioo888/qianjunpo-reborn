# 千軍破·重生 — Cocos Creator 客戶端 Shell

最小 **Cocos Creator 3.8.x** 工程：本地顯示 19×19 戰術棋盤、單位 Sprite／占位、**time_flow_rate** HUD。權威邏輯仍在 Go `pkg/tactical` / Roma；本工程僅 **顯示層**（白皮書邏輯／顯示分離）。

## 需求

- [Cocos Creator 3.8.4](https://www.cocos.com/creator-download)（3.8.x 皆可；`package.json` 標記 3.8.4）
- Mock 模式無需啟動 Janus／Roma；Live 模式需本機 `make compose-up` 或手動跑 Janus + Roma

## 開啟專案

1. 啟動 Cocos Dashboard → **新增／打開** → 選本目錄 `client/`（含 `package.json` 的資料夾）。
2. 首次開啟若提示升級引擎版本，選 **3.8.x** 並允許編譯 TypeScript。
3. 在 **資源管理器** 雙擊 `assets/scenes/TacticalBoard.scene`。
4. 點 **預覽**（瀏覽器或模擬器）：應看到 19×19 地格、兩枚單位（有貼圖時為 PX2D v02 Sprite）、左上角 HUD。

若場景腳本綁定遺失：在 `Canvas/TacticalRoot` 上 **添加组件 → 自定義脚本 → TacticalBootstrap**，`Snapshot Resource` 填 `data/tactical/demo_initial`。

## 戰棋操作（選取 → 高亮 → 移動）

顯示層互動由 **`TacticalBoardInteraction`**（與棋盤同節點）處理；合法格為 **BFS 可達格**（移動點數預設 **4**，對齊 `pkg/tactical` 決鬥單位；僅供高亮，權威仍以 Roma 驗證）。

| 操作 | 行為 |
|------|------|
| 點擊己方單位（`localPlayerId`，預設 **0**） | 黃框選取 + 綠色半透明合法目的地 |
| 點擊合法格 | 送出移動指令 |
| 點擊空白／非法格 | 取消選取 |

### Mock 驗證（預設）

1. `useLiveJanus = false`，`localPlayerId = 0`。
2. 預覽後點左側步兵（約格 `(2,8)`）。
3. 應出現綠色高亮格；點其中一格。
4. **預期**：單位移動、HUD 第三行 `Mock：已本地套用移動（非權威）`；瀏覽器 console 有 `[TacticalBootstrap] submit move`。
5. `lockstep frame` 在 HUD 第二行 +1（本地 mock 遞增，非 Roma）。

### Live 驗證

1. `make compose-up`，`useLiveJanus = true`，`liveBattleId = default/0`。
2. 選取與高亮同 Mock；點合法格後 **預期**：console 記錄指令；HUD 提示 **HTTP 指令鏡像尚未部署**（Janus 目前僅 `GET /v1/tactical/snapshot`）。
3. 權威提交請用 gRPC（與後端測試相同）：

   ```bash
   grpcurl -plaintext -d '{
     "battle_id":"default/0",
     "player_id":0,
     "kind":1,
     "unit_id":101,
     "to_x":5,
     "to_y":8
   }' localhost:9090 qianjunpo.gateway.v1.JanusGateway/SubmitTacticalCommand
   ```

4. 輪詢快照後棋面應與 Roma 一致；**TimeFlowHud** 仍只顯示 JSON 內 `timeFlowRateParts`／`lockstepFrame`。

### 相關腳本

| 檔案 | 職責 |
|------|------|
| `logic/ClientMoveReachability.ts` | 合法格 BFS（`CLIENT_DEFAULT_MOVE_POINTS=4`） |
| `logic/MockSnapshotMutator.ts` | Mock 本地套用移動 |
| `network/TacticalCommandClient.ts` | `submitTacticalMove`（POST 鏡像或 stub） |
| `display/TacticalBoardInteraction.ts` | 點選輸入 |
| `display/TacticalBoardView.ts` | 高亮層 + `pixelToGrid` |

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

2. **EnterBattle** 與 **GetBattleSnapshot** 回傳 `view_snapshot_json`（與 mock 同 schema）。Roma 會填入 **live** `timeFlowRateParts`（PR #17 後與 timedilation 一致）。

   ```bash
   grpcurl -plaintext -d '{"access_token":"dev","target_zone":{"zone_id":"default","shard":0}}' \
     localhost:9090 qianjunpo.gateway.v1.JanusGateway/EnterBattle
   grpcurl -plaintext -d '{"battle_id":"default/0"}' \
     localhost:9090 qianjunpo.gateway.v1.JanusGateway/GetBattleSnapshot
   ```

3. **Cocos 瀏覽器預覽**（HTTP 鏡像，與 gRPC 同源 JSON）：

   ```bash
   curl -s 'http://127.0.0.1:8090/v1/tactical/snapshot?battle_id=default/0' | jq '.timeFlowRateParts,.lockstepFrame'
   ```

   場景勾選 **`useLiveJanus`**，`liveBattleId` 填 `default/0`。Host／port 見 `JanusGatewayStub.ts` 的 `DEFAULT_NETWORK_STUB`。

### Live 輪詢與 HUD

- **`livePollIntervalMs`**（預設 **333ms**，約 3Hz；建議 **200–500ms**）：`useLiveJanus=true` 時以 `LiveViewSnapshotPoller` 週期呼叫 `fetchLiveViewSnapshot`；元件 `onDestroy` 會停止 timer。
- **`TimeFlowHudStub`**：每次成功快照呼叫 `updateFromSnapshot`，**僅**顯示 JSON 內 `timeFlowRateParts` 與 `lockstepFrame`（不插值、不造假速率）。
- 輪詢失敗（戰局不存在、網路錯誤）：HUD 第三行顯示狀態並 **指數退避** 重試（1s→10s cap）；成功後清除。
- **`TacticalBoardView`**：僅在幀／單位位置或 HP 變化時重繪單位層；地形成變才重繪棋盤。

### 驗證步驟（簡表）

| 步驟 | Mock | Live |
|------|------|------|
| 準備 | `make client-snapshot` | `make compose-up` + 可選 EnterBattle |
| Cocos | `useLiveJanus=false`，預覽 | `useLiveJanus=true`，`liveBattleId=default/0` |
| 預期 | HUD 顯示 mock 幀／速率；單位 v02 Sprite 或占位；左上資源圖標列（Nearest） | HUD 隨 curl 快照中 `timeFlowRateParts` 更新；失敗時 HUD 橙字狀態 |
| CLI | — | `curl -s 'http://127.0.0.1:8090/v1/tactical/snapshot?battle_id=default/0'` |

### 仍為 stub／後續

- Janus **TCP** framing（`:7000`）仍為 skeleton；正式客戶端應走 gRPC／未來 WebSocket。
- Cocos 內尚未內建 gRPC；Live 預覽用 HTTP **快照**鏡像；**SubmitTacticalCommand** 瀏覽器端為 stub（已嘗試 `POST /v1/tactical/command`，未部署則僅記錄 + HUD 提示）。
- `ViewSnapshot` 尚未帶每單位移動點數；高亮使用客戶端常數 4。
- Replay `TimeFlowRates` 串流尚未接入；目前僅 snapshot 內當幀 `timeFlowRateParts`。

## 目錄與分層

| 路徑 | 職責 |
|------|------|
| `assets/scripts/logic/` | 純資料：`ViewSnapshot`、合法格 BFS、Mock 移動 |
| `assets/scripts/display/` | 棋盤、點選互動、單位 Sprite／占位、HUD |
| `assets/scripts/network/` | 快照輪詢、`TacticalCommandClient` |
| `assets/scripts/app/` | `TacticalBootstrap` 場景入口 |
| `assets/resources/data/tactical/` | Mock 戰局 JSON |
| `assets/resources/textures/2d/` | 2D 像素貼圖（見目錄內 README） |
| `scripts/sync-wip-unit-textures.sh` | 自 `art/2d/_wip/units` 複製 v02（不改 art） |
| `scripts/sync-wip-icon-textures.sh` | 自 `art/2d/_wip/icons` 複製 13 枚 PX2D 圖標（不改 art） |

## 美術（卡通像素風）

- 單位：自 **`art/2d/_wip/units`** 複製 v02 至 `assets/resources/textures/2d/units/`（見 `textures/2d/README.md`）。
- 圖標：`bash client/scripts/sync-wip-icon-textures.sh` → `textures/2d/icons/`；執行期由 **`IconSpriteRegistry`** + **`PixelSpriteUtil`** 設 **Nearest**；戰術場景 HUD 下方有 **`ResourceIconHudStrip`**（五資源預覽，缺圖時 Sprite 關閉、不 crash）。
- Sprite **Nearest**、關 mipmap；棋盤格與 32px 圖標請 **整數倍** 縮放。
- 正式 `art/2d/**` 過審流程見 [`../art/2d/README.md`](../art/2d/README.md)；**勿**將 `_wip` 提升為正式夾。

## 建置產物

`library/`、`temp/`、`build/` 已在 `.gitignore`；請勿提交。

## 相關後端

- 戰術核心：`pkg/tactical`
- CLI：`cmd/match`、`make match-play`、`make client-snapshot`
- 閘道契約：`proto/gateway/v1/gateway.proto`、`proto/roma/v1/roma.proto`
- 整合測試：`make janus-roma-test`
