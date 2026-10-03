# 戰場預覽（field）

獨立靜態頁，公開路徑：**http://47.79.23.223/qjp/field/**（nginx `alias /var/www/qjp-static-preview/` 下之子目錄）。

- 不接 Janus／Live API、不用 `/qjp/` 根目錄既有卡通貼圖。
- 美術未到前：菱形色塊地格 + 色塊小兵堆 + 頭上兵力數字。

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
| 選取 | 戰場上 **點己方** 色塊堆（黃框） |
| 移動 | 選取後 **點綠色提示的空格**（曼哈頓距離 ≤4） |
| 攻擊 | 選取後 **點敵軍堆**（距離 ≤2）；只扣該支兵力，不會自動打完 |

敵方預設兩支（步、弓）固定不動。

## 部署（僅 field/）

**勿**執行 `deploy-web-preview.sh`（會 `--delete` 整包覆蓋 `/qjp/` 根）。

請使用同目錄上一層的 **`deploy-field-only.sh`**（只 rsync 本 `field/` 子目錄）。

### 證明 `/qjp/` 的 index 未被更換

部署腳本會在同步前後對 **`${TARGET_ROOT}/index.html`** 做 SHA-256；前後 digest 必須相同。若不同，腳本會失敗並提示。

手動複驗（在伺服器上）：

```bash
TARGET=/var/www/qjp-static-preview
sha256sum "$TARGET/index.html"   # 部署前記下
bash client/static-preview/deploy/deploy-field-only.sh "$TARGET"
sha256sum "$TARGET/index.html"   # 應與部署前一致
curl -sI "http://127.0.0.1/qjp/" | head -1
curl -sI "http://127.0.0.1/qjp/field/" | head -1
```

本次 PR **不在 CI／agent 上對 47.79.23.223 實際部署**。

## 檔案

| 檔案 | 說明 |
|------|------|
| `index.html` | 入口 |
| `field.css` | 樣式 |
| `field-main.js` | 選兵／畫布點擊 |
| `field-battle.js` | 戰鬥狀態與移動／攻擊 |
| `field-render.js` | Canvas 菱形格與單位 |
| `field-iso.js` | 斜角座標 |
