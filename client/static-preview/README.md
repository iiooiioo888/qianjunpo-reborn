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

子路徑部署時在 `index.html` 啟用：`<base href="/qjp/" />`。

## 本地開啟

```bash
cd client/static-preview
npx --yes serve . -l 3456
```

- Mock（預設）：<http://localhost:3456/> → `mock/demo_initial.json`（離線，僅本目錄即可）
- Live：<http://localhost:3456/?live=1>（需同域反代 `v1/lares/login` + `v1/tactical/*`，或 `?accessToken=` / `?mintUrl=`）
- 隱藏角色卡：`?cards=0`
- Mock 終局 overlay：`?mockVictory=win|lose|draw`（僅 mock，直接寫入快照 `winner`／`endReason` 欄位預覽）

僅部署本目錄（如 `deploy-web-preview.sh`）時不需 `client/assets`。

## Mock 同步

與 Cocos mock 同源更新：

```bash
cp client/assets/resources/data/tactical/demo_initial.json client/static-preview/mock/demo_initial.json
# 或：make client-snapshot 後再 cp
```

## 角色卡 v04（可選）

將 PNG 放到 `chars/*_v04.png`（檔名見 `config.js`）。缺圖自動占位，不影響棋盤。

## ISO25 地格（STANDARD）

草／山地貼圖由 `bash client/scripts/sync-wip-tile-textures.sh` 自 `art/25d/_wip/tiles` 複製至 `tiles/*.png`（檔名見 `config.js` `TERRAIN_TILE_SRC`）。缺圖時棋盤退回平面色塊。

## 互動

- 點己方單位（owner `0`）→ BFS 高亮（move **4**）。
- Mock：點高亮格 → 非權威 `applyMockMove`（與 Cocos mock 一致）。
- `?live=1`：啟動時 connect → enter-battle；點高亮格 → **同域** `POST v1/tactical/command`（帶 `session_id`＋`battle_id`）；**接受後** `POST v1/tactical/step-lockstep`（`steps: 4`，對齊 `CommandDelayFrames`）並套用回傳的 `view_snapshot_json`，再恢復快照輪詢；指令拒絕／HTTP／網路失敗顯示真實錯誤並可 **「重試戰術指令」**；建局／**快照輪詢**失敗可 **「重連 Live」**（或建局失敗時「重試 Live 建局」）。仍可用 `localDrift` mock 疊加。
- Live 技能：選己方單位後點 **「施放 stub 技能 (Strike)」**（射程內自動選敵格）或 **點敵方單位格** → `POST` `kind=5` + `skill_id`（預設 `1`，`?skillId=` 覆寫）→ step-lockstep → HUD **`lastSkillCast`** 與 HP 來自快照。

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

`timeFlowRateParts`、`lockstepFrame`、`sync: mock` / `sync: live`、`lastSkillCast`（技能執行後；缺欄位時保留上一筆 sticky）、終局 **棋盤 overlay**（`winner`／`endReason`，見 `docs/victory-live.md`）
