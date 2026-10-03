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

**accessToken**（三種方式，擇一）：

1. **推薦正式 URL**：<http://47.79.23.223/qjp/?live=1> — 無 query token 時會自動 `POST v1/lares/login`（**暫定契約**，與 `LaresAuth/Login` 對齊；核心 HTTP 鏡像合入後路徑可能微調）。
2. **一鍵**：自動 mint 失敗時點 HUD **「一鍵取 token」**（同上路徑）。
3. **手動**：`?live=1&accessToken=<token>`（Lares 簽發；勿對現網使用字面 `dev`）。

暫定 mint 預設 body：`{"username":"smoke","password":"smoke"}`（與 `docs/janus-http-mirror.md` smoke 帳密一致）。可用 `?mintUrl=`、`?mintUser=`、`?mintPass=` 覆寫。

子路徑部署時 `index.html` 會在 pathname 為 `/qjp` 或 `/qjp/…` 時自動插入 `<base href="/qjp/" />`；`paths.js` 亦會辨識 `/qjp` 掛載。

### Chrome 顯示 HTTP 400（根因）

線上若仍用 `return 301 /qjp/`（**無** `$is_args$args`），`/qjp?live=1` 會變成無 query 的 `/qjp/`。HTML 本身 curl 仍 **200**，但 Live 會打 `GET …/snapshot` **缺 `battle_id`** → Janus **400** `missing battle_id`（Chrome 網路面板／失敗請求像「整頁壞掉」）。**修法**：套用 `deploy/nginx-qjp-snippet.conf`（301 保留 query、`^~ /qjp/v1/` 優先於靜態、`try_files` 勿落到反代）。客戶端另備 `boot-query.js`（Referer 還原 query）與 `/qjp`→`/qjp/` 同步腳本，**不能替代** nginx 修正。

部署靜態與貼圖：

```bash
bash client/static-preview/deploy/deploy-web-preview.sh /var/www/qjp-static-preview
```

## 本地開啟

```bash
cd client/static-preview
npx --yes serve . -l 3456
```

- Mock（預設）：<http://localhost:3456/> → `mock/demo_initial.json`（離線，僅本目錄即可）
- Live：<http://localhost:3456/?live=1>（需同域反代 `v1/lares/login` + `v1/tactical/*`，或 `?accessToken=` / `?mintUrl=`）
- 隱藏角色卡：`?cards=0`
- Mock 終局 overlay：`?mockVictory=win|lose|draw`（僅 mock，直接寫入快照 `winner`／`endReason` 欄位預覽）。請用 **`/qjp/?mockVictory=win`**（斜線在 `?` 前），勿用 `/qjp?mockVictory=win`（易被 301 吃掉 query）。

僅部署本目錄（如 `deploy-web-preview.sh`）時不需 `client/assets`。

## Mock 同步

與 Cocos mock 同源更新：

```bash
cp client/assets/resources/data/tactical/demo_initial.json client/static-preview/mock/demo_initial.json
# 或：make client-snapshot 後再 cp
```

## 貼圖（deploy 必跑）

```bash
bash client/static-preview/scripts/sync-textures-from-assets.sh
```

自 `client/assets/resources/textures/2d/` 與 `art/2d/_wip/characters/`（v04 卡，**唯讀複製**）同步到 `textures/2d/`（`asset-registry.js`）。單位 `PX2D_unit_infantry`／`archer`／`cavalry`、地格 0–3、32px HUD／城鎮圖標、角色卡皆走此路徑；缺檔才色塊 fallback。

## ISO25 地格（STANDARD）

草／山／水／林地格貼圖由 `bash client/scripts/sync-wip-tile-textures.sh` 自 `art/25d/_wip/tiles` 複製至 `tiles/*.png`（檔名見 `config.js` `TERRAIN_TILE_SRC`）。缺圖時棋盤退回平面色塊。

## 互動（主線）

1. **排兵佈陣** → **補鎮／補給**（城鎮 32px 圖標）→ **開戰**（`campaign-flow.js`）。
2. `?live=1`：**開戰後** `connect` → `enter-battle` → 預設 **auto**（`live-auto-play.js`：`POST /v1/suggest` 或本地 AI → `POST v1/tactical/command` → `POST v1/tactical/step-lockstep` → `GET snapshot` 輪詢）直到 `winner`／`endReason` overlay（#94／#103）。`?auto=0` 關閉自動。
3. **手動點格移動／技能** 僅 `?debug=1`（或 `?debugMoves=1`）。`?skipCampaign=1` 跳過戰前步驟（測試用）。

### curl ↔ static-preview

| curl | 頁面 |
|------|------|
| `POST …/v1/lares/login`（暫定） | `live-gateway.js` `mintLiveAccessToken`（`?mintUrl=`） |
| `POST …/v1/tactical/connect` | `live-gateway.js`（`?connectUrl=`） |
| `POST …/v1/tactical/enter-battle` | 同上（`?enterBattleUrl=`） |
| `GET …/v1/tactical/snapshot?battle_id=…` | 預設 `liveUrl`（Enter 後 `battle_id`） |
| `POST …/v1/tactical/command` + Move body | 合法格點選（`?commandUrl=` 覆寫） |
| `POST …/v1/tactical/command` + `kind:5` + `skill_id` | 選單位 → 按鈕或點敵格（`stub-skill.js` / `app.js`） |
| `POST …/v1/tactical/step-lockstep` | 指令接受後 `live-gateway.js` / `app.js`（`?stepLockstepUrl=` 覆寫） |

Compose 直打範例見 [`../README.md`](../README.md) 與 [`../../docs/janus-http-mirror.md`](../../docs/janus-http-mirror.md)；Dev 公開頁用同域 `/qjp/v1/…` 或 `:18093/v1/…`，勿對外使用 `:18090`／`:18092`。

## HUD

`timeFlowRateParts`、`lockstepFrame`、`sync: mock` / `sync: live`、`lastSkillCast`（技能執行後；缺欄位時保留上一筆 sticky）、終局 **棋盤 overlay**（`winner`／`endReason`，見 `docs/victory-live.md`；Live 殲滅後若輪詢短暫回到 `endReason: "none"`，overlay 仍 sticky 至終局欄位，並停止輪詢）

### Live 殲滅（wipeout）驗證

| 步驟 | 指令／頁面 |
|------|------------|
| 核心 HTTP smoke | `WIPEOUT_SMOKE=1 ./scripts/janus-http-smoke.sh`（斷言 `endReason: "wipeout"`、`winner: 0\|1`） |
| 離線 overlay 契約 | `node client/scripts/static-preview-battle-end-smoke.mjs` |
| 瀏覽器 | `/qjp/?live=1` → Strike／移動至敵滅 → step-lockstep 後棋盤 overlay 顯示 **殲滅（全滅）** 與勝敗標題 |
