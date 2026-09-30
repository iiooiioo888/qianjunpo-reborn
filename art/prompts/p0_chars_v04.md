# P0 batch1 — 角色卡 AIGC 提示詞 v04

> **Historical only**：本檔 **不得** 作為生成依據。Canonical：**[`PROMPT_STANDARD_char_card_v1.md`](./PROMPT_STANDARD_char_card_v1.md)**。  
> 歷史鏈：v05 → v04 細化；v06 ultra 曾為過渡 canonical；現由 **STANDARD v1** 鎖定結構（範例見 STANDARD §5）。

## 硬門檻（全角色共用）

| 項目 | 規則 |
|------|------|
| 畫布 | **320×400 px** 精確（含透明邊亦不得改變外框尺寸） |
| 主體 | **單一角色**，無第二人物、無坐骑、無場景道具堆 |
| 背景 | 純色 **`#1A1410`**（無漸層、無紋理、無 vignette） |
| 線條 | 卡通像素、**1 px 黑色描邊**（outline），無抗鋸齒糊邊 |
| 構圖 | **正面半身**（shared front half-body pose）：肩線、手臂下垂角、鏡頭距離與 batch1 骨架一致（對齊 `p0_v03_pose_notes` 若存在） |

## 資產 ID

- `PX2D_CHAR_WEI_Caocao_ex`
- `PX2D_CHAR_SHU_Zhangfei_ex`
- `PX2D_CHAR_WU_Zhouyu_ex`（原吳占位升級為周瑜）

## 角色摘要（v04 簡表 — 僅存檔）

| 角色 | 必現 landmark |
|------|----------------|
| 曹操 | 黑翼平冠、細須、魏系深青綠袍與腰帶 |
| 張飛 | 大黑鬍、雙肩虎裘、蛇矛尖、蜀綠點綴 |
| 周瑜 | 紅絲巾、黑色文士高冠、紅金甲、年輕俊朗 |

## 正提示（v04 短版 — 已過時）

各角色請改 copy **`PROMPT_STANDARD_char_card_v1.md` §5 Examples**。
