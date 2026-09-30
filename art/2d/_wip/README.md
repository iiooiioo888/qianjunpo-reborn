# 2D WIP（P0）

本目錄下的 `characters/`、`units/`、`icons/` 為 **P0 工作稿**，非正式 production 美術。

- **風格**：卡通像素風；單一主體、shared front half-body pose
- **角色卡 batch1（identity v04）**：3 張 — `PX2D_CHAR_WEI_Caocao_ex_v04.png`、`PX2D_CHAR_SHU_Zhangfei_ex_v04.png`、`PX2D_CHAR_WU_Placeholder_01_v04.png`（見 `characters/`；皆 **320×400**）
- **單位 + 圖標 batch v02（cartoon-pixel）**：`units/` 步兵／騎兵 `*_v02.png`；`icons/` 13 張 `PX2D_icon_*_32.png`（**本批次未改動**）
- **取代說明**：v01／v02 角色卡（768×1024 級）及 PR #20 稿已 obsolete；角色以 **v04 identity** 為準（見 [`art/prompts/p0_batch1_chars_notes.md`](../../prompts/p0_batch1_chars_notes.md)）
- **來源**：Token Plan `wan2.7-image`（2026-09-30）
- **Endpoint**：`token-plan.cn-beijing`（使用 sk-sp- Token Plan key）
- **資產 ID**：與檔名一致（`PX2D_*` 前綴）
- **用途**：風格與 AIGC 管線驗證；過審後再遷至 `art/2d/characters/`、`art/2d/icons/` 等正式路徑

生成參數與請求紀錄：v01 POC 見 [`art/prompts/poc_v01_log.json`](../../prompts/poc_v01_log.json)、[`art/prompts/poc_v01_notes.md`](../../prompts/poc_v01_notes.md)；P0 batch1 角色 v04 見 [`art/prompts/p0_batch1_chars_notes.md`](../../prompts/p0_batch1_chars_notes.md)、[`art/prompts/p0_chars_v04_identity.md`](../../prompts/p0_chars_v04_identity.md)；P0 單位 v02 見 [`art/prompts/p0_units_v02.md`](../../prompts/p0_units_v02.md)。
