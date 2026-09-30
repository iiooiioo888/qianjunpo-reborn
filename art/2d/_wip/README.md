# 2D WIP（P0）

本目錄下的 `characters/`、`units/`、`icons/` 為 **P0 工作稿**，非正式 production 美術。

- **風格**：卡通像素風 **v02**（統一造型、單一主體、手部完整）
- **角色卡 batch1（v02）**：3 張 — `PX2D_CHAR_WEI_Caocao_ex_v02.png`、`PX2D_CHAR_SHU_Zhangfei_ex_v02.png`、`PX2D_CHAR_WU_Placeholder_01_v02.png`（見 `characters/`）
- **後續**：更多 **units**、**icons** 批次將依同一風格補齊
- **取代說明**：PR [#9](https://github.com/iiooiioo888/qianjunpo-reborn/pull/9) 的 POC 角色稿已過時；角色以本批 **v02** 為準（`units/`、`icons/` 仍為 v01 POC，待下一批更新）
- **來源**：Token Plan `wan2.7-image`（2026-09-30）
- **Endpoint**：`token-plan.cn-beijing`（使用 sk-sp- Token Plan key）
- **資產 ID**：與檔名一致（`PX2D_*` 前綴）
- **用途**：風格與 AIGC 管線驗證；過審後再遷至 `art/2d/characters/`、`art/2d/icons/` 等正式路徑

生成參數與請求紀錄：v01 POC 見 [`art/prompts/poc_v01_log.json`](../../prompts/poc_v01_log.json)、[`art/prompts/poc_v01_notes.md`](../../prompts/poc_v01_notes.md)；P0 batch1 角色 v02 見 [`art/prompts/p0_batch1_chars_notes.md`](../../prompts/p0_batch1_chars_notes.md)。
