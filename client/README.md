# 千軍破·重生 — Cocos Creator 客戶端 Shell

最小 **Cocos Creator 3.8.x** 工程：本地顯示 19×19 戰術棋盤、單位占位、**time_flow_rate** HUD 掛點。權威邏輯仍在 Go `pkg/tactical` / Roma；本工程僅 **顯示層**（白皮書邏輯／顯示分離）。

## 需求

- [Cocos Creator 3.8.4](https://www.cocos.com/creator-download)（3.8.x 皆可；`package.json` 標記 3.8.4）
- 無需啟動 Janus／Roma 即可預覽棋盤（mock JSON）

## 開啟專案

1. 啟動 Cocos Dashboard → **新增／打開** → 選本目錄 `client/`（含 `package.json` 的資料夾）。
2. 首次開啟若提示升級引擎版本，選 **3.8.x** 並允許編譯 TypeScript。
3. 在 **資源管理器** 雙擊 `assets/scenes/TacticalBoard.scene`。
4. 點 **預覽**（瀏覽器或模擬器）：應看到 19×19 地格（河／關隘橋）、兩枚占位單位、左上角 `time_flow_rate`／幀號 stub。

若場景腳本綁定遺失：在 `Canvas/TacticalRoot` 上 **添加组件 → 自定義脚本 → TacticalBootstrap**，`Snapshot Resource` 填 `data/tactical/demo_initial`。

## 目錄與分層

| 路徑 | 職責 |
|------|------|
| `assets/scripts/logic/` | 純資料：`ViewSnapshot` 解析（無權威、無 Cocos 依賴） |
| `assets/scripts/display/` | 棋盤／單位占位／色板／HUD stub |
| `assets/scripts/network/` | `JanusGatewayStub`（後續接 `proto/gateway` TCP） |
| `assets/scripts/app/` | `TacticalBootstrap` 場景入口 |
| `assets/resources/data/tactical/` | Mock 戰局 JSON |
| `assets/resources/textures/2d/` | 正式 2D 像素貼圖匯入位（見 [`../art/2d/README.md`](../art/2d/README.md)） |

## Mock 資料（與後端對齊）

預設 JSON 由 Go 匯出，對應 `tactical.NewMatch(0xcafe)` 初始局面（與 `make match-play` 同 seed）：

```bash
# 在倉庫根目錄
make client-snapshot
# 或
go run ./cmd/client-snapshot -seed 0xcafe -out client/assets/resources/data/tactical/demo_initial.json
```

JSON schema：`schemaVersion=1`，`boardSize=19`，`cells[y][x]` 地形與 `pkg/board` 枚舉一致，`timeFlowRateParts` 為萬分比（10000=1.0x，見 `pkg/timedilation`）。

## 後續：即時連線（非本 PR 範圍）

1. Janus `JanusGateway.Connect` / `EnterBattle`（`proto/gateway/v1/gateway.proto`）。
2. 週期拉取或推送 **view snapshot**（或由 lockstep frame + 指令在客戶端 replay 純顯示層 — 仍不得本地改 HP）。
3. 將 `TimeFlowHudStub.setTimeFlowRateParts` 接到心跳或 replay v2 `TimeFlowRates`。
4. 替換 `JanusGatewayStub` 為 TCP/WebSocket 實作（參考 `services/janus` 與 `make janus-roma-test`）。

## 美術（卡通像素風）

- 正式資產：`art/2d/**` 過審後拷貝至 `assets/resources/textures/2d/`；匯入規範見 [`art/2d/README.md`](../art/2d/README.md)。
- 目前棋盤為 **Graphics 色塊 + 占位幾何**；接圖後改 `TacticalBoardView` / `UnitPlaceholderView` 即可，logic 層不變。

## 建置產物

`library/`、`temp/`、`build/` 已在 `.gitignore`；請勿提交。

## 相關後端

- 戰術核心：`pkg/tactical`
- CLI：`cmd/match`、`make match-play`
- 閘道契約：`proto/gateway/v1/gateway.proto`
