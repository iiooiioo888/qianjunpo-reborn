# 2D WIP（P0）

本目錄下的 `characters/`、`units/`、`icons/` 為 **P0 工作稿**，非正式 production 美術。

- **風格**：卡通像素風 **v02**（統一造型、單一主體、手部完整）
- **v03（pose／比例鎖）**：正面半身 **相同骨架**（肩／臂／鏡頭一致），頭身比 **1:2.2–1:2.5**；角色 `characters/*_v03.png` 供姿勢審閱並**取代 v02 作 pose review 基準**（v02 仍保留）；單位 `units/*_v03.png` 與 v02 **並存**對照
- **角色卡 batch1（v02）**：3 張 — `PX2D_CHAR_WEI_Caocao_ex_v02.png`、`PX2D_CHAR_SHU_Zhangfei_ex_v02.png`、`PX2D_CHAR_WU_Placeholder_01_v02.png`（見 `characters/`）
- **角色卡 batch1（v03，pose-aligned）**：同上 3 角色之 `*_v03.png`（見 `characters/`）
- **單位 + 圖標 batch v02（cartoon-pixel）**：`units/` 新增步兵／騎兵 `*_v02.png`；`icons/` 13 張 `PX2D_icon_*_32.png` 已更新為同批 v02 風格（檔名無 `_v02` 後綴）
- **單位 v03（pose-aligned）**：`PX2D_unit_infantry_v03.png`、`PX2D_unit_cavalry_v03.png`（共用 idle 站姿骨架）
- **取代說明**：PR [#9](https://github.com/iiooiioo888/qianjunpo-reborn/pull/9) 的 POC 角色稿已過時；角色以 batch1 **v02** 為準。`units/`、`icons/` 的 v01 仍保留作對照，審閱以 v02 單位與現行 `icons/` 為準
- **來源**：Token Plan `wan2.7-image`（2026-09-30）
- **Endpoint**：`token-plan.cn-beijing`（使用 sk-sp- Token Plan key）
- **資產 ID**：與檔名一致（`PX2D_*` 前綴）
- **用途**：風格與 AIGC 管線驗證；過審後再遷至 `art/2d/characters/`、`art/2d/icons/` 等正式路徑

生成參數與請求紀錄：v01 POC 見 [`art/prompts/poc_v01_log.json`](../../prompts/poc_v01_log.json)、[`art/prompts/poc_v01_notes.md`](../../prompts/poc_v01_notes.md)；P0 batch1 角色 v02 見 [`art/prompts/p0_batch1_chars_notes.md`](../../prompts/p0_batch1_chars_notes.md)；P0 單位 v02 見 [`art/prompts/p0_units_v02.md`](../../prompts/p0_units_v02.md)；P0 v03 pose lock 見 [`art/prompts/p0_v03_pose_notes.md`](../../prompts/p0_v03_pose_notes.md)。
