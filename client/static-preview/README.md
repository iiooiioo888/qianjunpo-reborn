# static-preview（《千軍破》戰術棋盤）

輕量 **HTML + Canvas + ES module** 預覽頁，讀取與 Cocos 相同的 `ViewSnapshot` JSON（`demo_initial` schema）。**不需要 Cocos Creator**。

## 本地開啟

必須用 HTTP 服務（`file://` 無法 `fetch` mock JSON，且 Live 會受 CORS 限制）。

```bash
cd client
npx --yes serve . -l 3456
```

瀏覽器：

- Mock（預設）：<http://localhost:3456/static-preview/>
- Live Janus 鏡像：<http://localhost:3456/static-preview/?live=1>
- 自訂 Live URL：<http://localhost:3456/static-preview/?live=1&liveUrl=https%3A%2F%2Fexample%2Fjanus%2Fv1%2Ftactical%2Fsnapshot%3Fbattle_id%3Ddefault%2F0>
- 隱藏角色卡：<http://localhost:3456/static-preview/?cards=0>

Mock 資料路徑（相對本目錄）：`../assets/resources/data/tactical/demo_initial.json`  
角色卡 v04 PNG：`../assets/resources/textures/2d/chars/*_v04.png`（需從 `client/` 根目錄 serve）。

### 互動（mock / 本地漂移）

- 點選己方單位（owner `0`）→ 以 BFS 高亮可走格（move 預算 **4**，8 邻格）。
- 再點高亮格 → **非權威** 本地 mock 移動（僅預覽）。
- `?live=1` 時預設只讀拉快照；若仍做本地移動會標記 `sync: live · stale`。

## Live 快照（開發機）

預設 URL（可在 `config.js` 或 `?liveUrl=` 覆寫）：

`http://47.79.23.223:18090/v1/tactical/snapshot?battle_id=default/0`

Janus HTTP 鏡像**未保證**對任意網頁開 CORS。從瀏覽器直連 `:18090` 可能失敗。建議由 **Dev nginx 同域反代**，例如：

```nginx
# 範例片段（由 Dev 部署，非本倉庫 CI）
location /janus/ {
    proxy_pass http://127.0.0.1:18090/;
    add_header Access-Control-Allow-Origin *;
}
```

頁面則使用：`?live=1&liveUrl=/janus/v1/tactical/snapshot?battle_id=default%2F0`

## 預期 Dev 掛載（待 ops）

靜態檔建議由 nginx 對外：

- **URL**：<http://47.79.23.223/qjp/> → 本目錄 `client/static-preview/`（或整個 `client/` 下 `/qjp/` 子路徑）
- 埠號可不為 80；路徑以 Dev 環境為準。

部署與 TLS／反代由 Dev 另行配置；本目錄僅提供可 serve 的靜態資源。

## HUD

- `timeFlowRateParts`（萬分比，10000 = 1.0×）
- `lockstepFrame` + `sync: mock` / `sync: live` 行（對齊 Cocos `LockstepHudFormat` 文案）
