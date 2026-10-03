# 戰場預覽（field）

獨立靜態頁，公開路徑：**http://47.79.23.223/qjp/field/**（nginx `alias /var/www/qjp-static-preview/` 下之子目錄）。

- 不接 Janus／Live API、不用 `/qjp/` 根目錄既有卡通貼圖。
- 手繪稿僅放在本目錄；地格、兵堆、選兵面板為 PNG，**兵力數字由程式繪製**（原稿頂部約 y&lt;56 的佔位字在繪製時裁切，不改 PNG）。

## 手繪資源（入庫 MD5）

| 檔案 | 尺寸 | MD5 |
|------|------|-----|
| `QJP_ground_grass_v01.png` | 256×128 | `6aa246ae96bda4b91bd186572277e729` |
| `QJP_stack_infantry_v01.png` | 192×224 | `0d9c45e9bc692044609298a4ca797265` |
| `QJP_stack_archer_v01.png` | 192×224 | `3663ebb7131bd2dc547abe8251677f84` |
| `QJP_stack_cavalry_v01.png` | 192×224 | `ec61a124d4327073c22accc3974f765f` |
| `QJP_panel_teal_gold_v01.png` | 400×160 | `9bec886b33333ee0481b3ef685b8edcb` |

## 本地預覽

```bash
cd client/static-preview/field
npx --yes serve . -l 3460
# http://localhost:3460/
```

或從 `client/static-preview` 根目錄 serve，開 **http://localhost:3456/field/**。

## 操作

| 階段 | 操作 |
|------|------|
| 選兵 | 拖曳步／弓／騎兵力滑桿 → 按 **「出征進戰場」** |
| 選取 | 戰場上 **點己方** 兵堆（黃框） |
| 移動 | 選取後 **點綠色提示的空格**（曼哈頓距離 ≤4） |
| 攻擊 | 選取後 **點敵軍堆**（距離 ≤2）；只扣該支兵力，不會自動打完 |
| 勝負 | 敵方兩支兵力合計歸零 → 畫布 **「勝」** 並鎖住點擊；己方三支合計歸零 → **「負」** 並鎖住。兵力歸零的單位離場（不繪製、不可點） |

敵方預設兩支（步、弓）固定不動、不反擊。

## 部署（僅 field/）

**勿**執行 `deploy-web-preview.sh`（會 `--delete` 整包覆蓋 `/qjp/` 根）。

請使用同目錄上一層的 **`deploy-field-only.sh`**（只 rsync 本 `field/` 子目錄）。

### 證明 `/qjp/` 的 index 未被更換

部署腳本會在同步前後對 **`${TARGET_ROOT}/index.html`** 做 SHA-256；前後 digest 必須相同。若不同，腳本會失敗並提示。

本次 PR **不在 CI／agent 上對 47.79.23.223 實際部署**。

## 檔案

| 檔案 | 說明 |
|------|------|
| `index.html` | 入口 |
| `field.css` | 樣式 |
| `field-main.js` | 選兵／畫布點擊 |
| `field-battle.js` | 戰鬥狀態與移動／攻擊 |
| `field-render.js` | Canvas 菱形格與單位 |
| `field-assets.js` | 手繪 PNG 載入 |
| `field-iso.js` | 斜角座標與點選（兵身優先，否則菱形） |
| `field-stack-layout.js` | 兵堆繪製與點選共用版面 |
| `QJP_*.png` | 手繪貼圖（見上表） |
