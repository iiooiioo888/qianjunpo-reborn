# 2D 卡通像素風 → Cocos 客戶端匯入

本目錄為 **已過審** 的 2D 正式資產歸檔位（見 [`../README.md`](../README.md)）。PR #9 等 pure-pixel POC 稿僅作管線參考，**勿**直接當 production。

## 美術方向（顯示層）

- **卡通像素風**：誇張輪廓 + 硬邊像素、固定比例、**1px 外描邊**、左上光源、共用調色盤、圖標對齊 **32px 網格**。
- Cocos 中 Sprite 使用 **Nearest**、僅 **整數倍** 縮放（與 `art/README.md` 2D 章節一致）。

## 匯入流程

1. 美術在 Notion 過審後，將 PNG 自 `art/2d/_wip/` **移入** 對應正式子目錄：
   - `art/2d/characters/` — 角色卡
   - `art/2d/icons/` — 資源／建築／技能圖標
   - `art/2d/ui/` — 面板與框線
   - （戰場單位 sprite 建議新增 `art/2d/units/`，與 contentops ID 對齊後登記 Notion）
2. **複製或 symlink** 至 Cocos 工程：
   - 目標根：`client/assets/resources/textures/2d/`
   - 子目錄鏡像：`characters/`、`icons/`、`units/`、`ui/`
3. 在 Cocos **資源管理器** 中選中目錄 → 右鍵 **重新導入**；確認 texture 的 `filterMode` 為 **Nearest**。
4. 顯示層腳本（`client/assets/scripts/display/`）替換 `UnitPlaceholderView` / 地格色塊為 SpriteFrame；**邏輯層**仍只讀 `ViewSnapshot` JSON，不持有權威 HP／座標來源。
5. 可選：依 [`docs/aigc/PIPELINE.md`](../../docs/aigc/PIPELINE.md) 同步 manifest 至 `assets/import/`（引擎 bundle 執行層）。

## 檔名

與 Notion 資產 ID 一致（例：`PX2D_unit_infantry_v01.png` → `textures/2d/units/`）。過審前的工作稿留在 `_wip`，不得進本目錄鏈路。

## 相關

- 客戶端工程說明：[`client/README.md`](../../client/README.md)
- 戰術 mock 資料：`make client-snapshot` → `client/assets/resources/data/tactical/demo_initial.json`
