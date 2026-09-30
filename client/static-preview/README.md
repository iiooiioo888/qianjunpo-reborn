# static-preview（《千軍破》戰術棋盤）

**目錄（寫死）：** `client/static-preview/`  
輕量 HTML + Canvas + ES module，`ViewSnapshot`（`demo_initial` schema）。**非 Cocos web-mobile**。

## 正式／備用公開 URL（Dev）

| 用途 | URL |
|------|-----|
| **正式** | <http://47.79.23.223/qjp/>（`location /qjp/`，不佔 Linkin `:80` 根） |
| **備用** | <http://47.79.23.223:18093/>（同機 Janus 靜態＋API 根） |

**勿用 `:18092`** 作為對外預覽埠。Live 快照一律走**同域相對** `v1/tactical/snapshot?battle_id=default/0`：

- 掛在 `/qjp/` → `/qjp/v1/tactical/snapshot?battle_id=default/0`
- 掛在 `:18093` 根 → `/v1/tactical/snapshot?battle_id=default/0`

Dev nginx 已將上述路徑反代到 Janus。`?live=1` 會先 **`POST v1/tactical/connect`** → **`POST v1/tactical/enter-battle`**（#60），再輪詢快照／發指令；可用 `?connectUrl=`、`?enterBattleUrl=` 覆寫路徑。

**accessToken**：預覽須帶 Lares 簽發 token，例如 `?live=1&accessToken=<token>`（預設空，勿對現網使用字面 `dev`）。缺 token 時狀態列提示並可重試建局。

子路徑部署時在 `index.html` 啟用：`<base href="/qjp/" />`。

## 本地開啟

```bash
cd client/static-preview
npx --yes serve . -l 3456
```

- Mock（預設）：<http://localhost:3456/> → `mock/demo_initial.json`（離線，僅本目錄即可）
- Live：<http://localhost:3456/?live=1&accessToken=…>（本地需自行反代 `v1/` 或 `?liveUrl=` 指到可達端點）
- 隱藏角色卡：`?cards=0`

僅部署本目錄（如 `deploy-web-preview.sh`）時不需 `client/assets`。

## Mock 同步

與 Cocos mock 同源更新：

```bash
cp client/assets/resources/data/tactical/demo_initial.json client/static-preview/mock/demo_initial.json
# 或：make client-snapshot 後再 cp
```

## 角色卡 v04（可選）

將 PNG 放到 `chars/*_v04.png`（檔名見 `config.js`）。缺圖自動占位，不影響棋盤。

## 互動

- 點己方單位（owner `0`）→ BFS 高亮（move **4**）。
- Mock：點高亮格 → 非權威 `applyMockMove`（與 Cocos mock 一致）。
- `?live=1`：啟動時 connect → enter-battle；點高亮格 → **同域** `POST v1/tactical/command`（帶 `session_id`＋`battle_id`）；成功後輪詢快照；建局／**快照輪詢**失敗顯示真實 HTTP 並可一鍵 **「重連 Live」**（或建局失敗時「重試 Live 建局」）。仍可用 `localDrift` mock 疊加。

### curl ↔ static-preview

| curl | 頁面 |
|------|------|
| `POST …/v1/tactical/connect` | `live-gateway.js`（`?connectUrl=`） |
| `POST …/v1/tactical/enter-battle` | 同上（`?enterBattleUrl=`） |
| `GET …/v1/tactical/snapshot?battle_id=…` | 預設 `liveUrl`（Enter 後 `battle_id`） |
| `POST …/v1/tactical/command` + Move body | 合法格點選（`?commandUrl=` 覆寫） |

Compose 直打範例見 [`../README.md`](../README.md) 與 [`../../docs/janus-http-mirror.md`](../../docs/janus-http-mirror.md)；Dev 公開頁用同域 `/qjp/v1/…` 或 `:18093/v1/…`，勿對外使用 `:18090`／`:18092`。

## HUD

`timeFlowRateParts`、`lockstepFrame`、`sync: mock` / `sync: live`
