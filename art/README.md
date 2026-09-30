# 美術資產目錄（`art/`）

《千軍破·重生》美術產出與最終遊戲資產的 **Git 歸檔根目錄**。依製作管線分為 **2D（像素風）**、**2.5D（等距沙盤）**、**3D（低面演出）** 三條線；AIGC 生成稿與人工修稿皆先進本倉庫，**過審後**才進入對應正式子目錄並合入 `main`。

> 產線與 ComfyUI／worker 對接說明見 [`docs/aigc/PIPELINE.md`](../docs/aigc/PIPELINE.md)。本目錄負責 **檔案落位、命名與審核狀態**；`assets/import/` 仍為引擎匯入與 manifest 的執行層（見 PIPELINE「Asset import path」）。

Notion 美術專區（任務、風格板、審核紀錄）：[美術首頁](https://app.notion.com/p/3eb3a44d333d81f6ba0dc8ab43e20164)（🎨 千軍破美術）；文檔總覽見 [Notion 文檔總覽](https://app.notion.com/p/3eb3a44d333d8114a5dbd891c9e831b4)。檔名與資產 ID 須與本 README 及 Notion 登記一致。

---

## 目錄結構

| 路徑 | 用途 |
|------|------|
| `art/2d/characters/` | 像素角色卡、戰場單位立繪／sprite |
| `art/2d/ui/` | 像素 UI 圖標、面板裝飾、框線 |
| `art/2d/icons/` | 技能、資源、狀態等小圖標 |
| `art/25d/tiles/` | 等距地格、地形貼圖 |
| `art/25d/roads/` | 道路、橋、邊界拼接 |
| `art/25d/buildings/` | 等距建築與佔格物件 |
| `art/3d/units/` | 低面單位模型匯出（遊戲用） |
| `art/3d/anims/` | 動畫片段、骨骼動畫匯出 |
| `art/3d/vfx/` | 粒子、材質、VFX 匯出 |
| `art/prompts/` | `prompt_log`、生成參數 JSON（可選，與資產 ID 對應） |
| `*/_wip/` | **未過審** 的生成稿與迭代稿（見下文） |

空目錄以 `.gitkeep` 佔位，避免 Git 忽略目錄本身。

---

## 命名規範（對齊資產 ID）

檔名 **必須** 與內容營運／Notion 上的資產 ID 一致，方便 `pkg/contentops` 審核與 manifest 登錄。

### 通用規則

- 僅使用 **小寫英數與底線** `a-z0-9_`（擴展名除外）。
- 版本尾碼：`v01`、`v02` …（同一 ID 換稿時遞增，勿覆蓋已合入 `main` 的檔案）。
- 範例：`ISO25_tile_grass_v01.png`、`CHAR_zhaoyun_card_v03.png`。

### 前缀建議

| 前缀 | 範例 | 說明 |
|------|------|------|
| `CHAR_` | `CHAR_liubei_unit_v01.png` | 2D 角色／單位 |
| `UI_` | `UI_panel_frame_v01.png` | 2D UI |
| `ICO_` | `ICO_skill_fire_v01.png` | 2D 圖標 |
| `ISO25_` | `ISO25_building_barracks_v01.png` | 2.5D 等距資產 |
| `MDL_` | `MDL_cavalry_low_v01.glb` | 3D 匯出模型 |
| `VFX_` | `VFX_arrow_trail_v01.png` | 3D／特效相關貼圖或序列 |

完整 ID 登記表以 [美術首頁](https://app.notion.com/p/3eb3a44d333d81f6ba0dc8ab43e20164) 為準；新增前缀前請在 Notion 登記，避免與程式內容鍵衝突。

---

## 2D 像素（`art/2d/`）

- **縮放**：遊戲內使用 **Nearest（最近鄰）** 濾鏡，僅允許 **整數倍** 放大（1×、2×、3×…），禁止非整數縮放導致糊化。
- **角色卡**：約 **320×400** px（可含透明邊；以 Notion 版型為準）。
- **戰場單位**：立繪或 sprite 高度建議 **64–96** px（與地格視覺比例一致）。
- **格式**：主交付 `PNG`（含 alpha）；動畫可用 `PNG` 序列或團隊約定的 sprite sheet 命名（sheet 名仍對齊資產 ID）。

---

## 2.5D 等距沙盤（`art/25d/`）

- **地格尺寸**：**64×32** px，**2:1 等距**（菱形 tile，寬為高的 2 倍）。
- **拼接**：tile／road／building 須提供清晰 anchor 與鄰接邊，檔名區分地形與等級（如 `ISO25_tile_grass_v01.png`）。
- **縮放**：同 2D，顯示層建議整數倍；源稿保持像素對齊。

---

## 3D 低面演出（`art/3d/`）

- **風格**：低面（low-poly）、風格化材質，面數與貼圖尺寸以 Notion 技術規格為準。
- **源檔與匯出分離**：
  - **勿將** `.blend`、`.max`、`.psd` 等大體積源檔提交至本倉庫（放團隊物件儲存或 Notion 附件連結）。
  - **`art/3d/` 內只放遊戲用匯出**（如 `.glb`、`.gltf`、烘焙貼圖、動畫片段）；若需記錄源檔位置，在 `art/prompts/` 或 Notion 該資產頁註明路徑與版本。
- **命名**：匯出檔使用 `MDL_`／`VFX_` 等前缀，與 2D／2.5D 同一套 ID 體系。

---

## `_wip` 與審核流程

1. **AIGC 或草稿** 一律先放入對應管線的 `art/<管線>/_wip/`（或 `art/prompts/` 記錄參數）。
2. **未過審** 的生成圖 **不得** 直接放入 `characters/`、`tiles/` 等正式目錄。
3. 美術／營運在 [美術首頁](https://app.notion.com/p/3eb3a44d333d81f6ba0dc8ab43e20164) 標記通過後：
   - 將檔案 **移動**（非僅複製）到正式目錄並確認檔名符合資產 ID；
   - 刪除或清空 `_wip` 中對應舊稿，避免重複 ID；
   - 依 [`docs/aigc/PIPELINE.md`](../docs/aigc/PIPELINE.md) 必要時再匯出至 `assets/import/` 並更新 manifest。
4. **合入 `main`**：僅包含 **已過審** 的正式目錄資產；含 `_wip` 的 PR 限開發分支，合入前須清空或移出未審內容（與 contentops 審核一致）。

---

## `art/prompts/`（可選）

- 建議檔名：`<資產ID>_prompt.json` 或 `<資產ID>_prompt_log.md`。
- 記錄：模型、seed、ComfyUI workflow 節點摘要、參考圖 ID；**勿** 提交模型權重或 API 密鑰。
- 與 [`docs/aigc/PIPELINE.md`](../docs/aigc/PIPELINE.md) 中的 worker／ComfyUI 流程對齊，便於重現與稽核。

---

## 相關文件

- [AIGC 產線 stub](../docs/aigc/PIPELINE.md) — worker、ComfyUI、匯入 `assets/import/`
- [Notion 美術首頁](https://app.notion.com/p/3eb3a44d333d81f6ba0dc8ab43e20164) — 風格、任務、審核狀態（主資料源）
- [Notion 文檔總覽](https://app.notion.com/p/3eb3a44d333d8114a5dbd891c9e831b4) — 項目文檔索引
