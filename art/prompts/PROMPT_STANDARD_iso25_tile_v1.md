# PROMPT STANDARD — ISO25 地格 tile AIGC v1（**canonical · locked**）

> 64×32 戰場 **2:1 isometric diamond** 地格樣張。**結構對齊** unit STANDARD（A→E 順序）；僅 **硬輸出尺寸**、**鑽石 footprint**、**場地透明角** 與 unit／角色卡不同。  
> **禁止** 交付正方形不透明角（#28 lesson：opaque square corners = fail）。修訂本檔或增補 Example；勿平行開規格檔。

---

## 1. Purpose · 目的

wan2.7-image 生成 **ISO25** 卡通像素地格（如 grass／mountain／water／forest），與 2D track **同一 STYLE LOCK 調色語言**（`#8B4513` / `#CD5C5C` / `#DAA520` + terrain greens／greys），經 post-process 裁成精確 **64×32** 並套鑽石 alpha mask，供 art-manager 樣張過審。

---

## 2. Pipeline · 流程

填模板 → 單条 positive（E 段 Forbidden 內嵌）→ API（通常 `1024*1024`）→ **NN resize 64×32** → **鑽石外 alpha=0** → 驗證四角 alpha=0 → Notion → art-manager。

---

## 3. Sections A→E（tile 專用差異）

| 段 | Tile 規則 |
|----|-----------|
| **A. HARD OUTPUT** | Exact **64×32**（API 可先大圖再 post），**one subject**，**2:1 isometric diamond** footprint；**outside diamond MUST alpha=0**（四角透明）；no sheet/collage/gold frame/text；solid dark or transparent-ready backdrop outside diamond |
| **B. STYLE LOCK** | Cartoon pixel, **1px hard edge**, no AA, top-left light；palette **`#8B4513` / `#CD5C5C` / `#DAA520`** 小面積點綴 + **terrain greens／greys** 為主 |
| **C. POSE LOCK** | **Flat sand-table tile**（非單位全身）；鑽石佔滿畫幅；草地＝平鋪可走；山地＝rocky／highland **impassable readable**；水域＝水面不可／難通；樹林＝可走覆蓋感（勿做成石山） |
| **D. IDENTITY SLOT** | MUST SHOW FIRST：地類輪廓（grass tufts／rock peak）→ 主色塊 → 邊緣可讀 → 通行語意（passable vs impassable） |
| **E. FORBIDDEN** | opaque square corners + wrong size not 64x32 + multi-tile sheet + collage + text + gold frame + photoreal + buildings／units on tile（除非 Example 明確） |

---

## 4. Empty template · 空白模板

```
A. HARD OUTPUT: exact 64x32 pixel isometric terrain tile, single subject only, 2:1 isometric diamond footprint filling canvas, outside diamond transparent alpha zero, no sprite sheet no collage no gold frame no text,
B. STYLE LOCK: cartoon pixel art, 1px hard black outline, no anti-aliasing, top-left light, palette terrain {{TERRAIN_COLORS}} accents #8B4513 #CD5C5C #DAA520,
C. POSE LOCK: sand-table flat isometric tile centered diamond, {{TILE_HEIGHT_HINT}},
D. IDENTITY {{TILE_ID}} MUST SHOW FIRST: (1) silhouette {{SILHOUETTE}}, (2) main color blocks {{COLORS}}, (3) edge readability {{EDGE}}, (4) passability cue {{PASS}},
E. FORBIDDEN: opaque square corners wrong size not 64x32 photoreal gradient sprite sheet collage text gold frame buildings units {{TILE_FORBIDDEN}}
```

---

## 5. Examples · 範例（非新版本）

見同目錄：

- `ISO25_tile_grass_v02.txt`
- `ISO25_tile_mountain_v01.txt`
- `ISO25_tile_water_v01.txt`
- `ISO25_tile_forest_v01.txt`

---

## 6. Post-process hard gate（必須）

1. Nearest-neighbor → **exact 64×32**
2. Diamond vertices：**mid-top, mid-right, mid-bottom, mid-left**；外像素 **RGBA alpha=0**
3. Self-check：`WxH==64x32` 且四角 alpha==0；否則 reject（#28）

---

## 7. Historical

- `grass_v01` 等舊件四角不透明 → **不可作交付**；以本 STANDARD + mask 為準。
