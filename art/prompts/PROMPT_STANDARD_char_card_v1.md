# PROMPT STANDARD — 角色卡 AIGC v1（**canonical · locked**）

> **唯一生成規格**：所有《千軍破·重生》2D 角色卡（320×400）AIGC 提示詞 **必須** 依本檔結構組裝。  
> **禁止** 再新增 `p0_chars_v07_*` 等版本檔；僅可 **修訂本 STANDARD** 或增補 **IDENTITY 範例**（見 §6）。  
> 歷史稿：[`p0_v03_pose_notes.md`](./p0_v03_pose_notes.md) · [`p0_chars_v04.md`](./p0_chars_v04.md) · [`p0_chars_v05_detail.md`](./p0_chars_v05_detail.md) · [`p0_chars_v06_ultra.md`](./p0_chars_v06_ultra.md) — **僅供對照，不可作生成依據**。

---

## 1. Purpose · 目的

為 **wan2.7-image**（及同構 multimodal 端點）提供 **單一、不可拆段** 的提示詞結構：每次生成 = 依序填滿 **A→E** 五段並合成 **一條 positive string**（負向約束 **寫入 positive 內**，不另建 negative 欄位，除非 API 強制）。

- **不變量**：畫布、風格、姿勢骨架、Forbidden 尾段  
- **變量**：**D. IDENTITY SLOT**（角色 landmark／陣營色律）  
- **輸出**：`PX2D_CHAR_<FACTION>_<Name>_ex` 對應 WIP PNG → Notion 登記 → art-manager 審核

---

## 2. Pipeline · 流程

| 步驟 | 動作 |
|------|------|
| 1 | 自 §4 **空白模板** 複製，填入 D 段 IDENTITY（必須 **MUST SHOW FIRST** 順序：帽/臉/服色/武器或道具/陣營色律） |
| 2 | 合併為 **單一 positive**（A+B+C+D+E 連續一段，建議英文關鍵詞 + 必要中文 landmark） |
| 3 | 呼叫 API：`model: wan2.7-image`（見 [`poc_v01_notes.md`](./poc_v01_notes.md)）；記錄 seed／參數至 `art/prompts/` 可選 log |
| 4 | **驗證硬輸出**：精確 **320×400** px、背景 **`#1A1410`** 取樣、單主體、無 sheet／拼貼／框／字 |
| 5 | 上傳 **Notion** 美術資產頁（ID 與檔名一致） |
| 6 | **art-manager** 對照 D 段 landmark + §5 範例級細節審圖；通過後移出 `_wip` |

---

## 3. Template sections · 段落順序（每條 prompt string **必須** 含 A→E）

生成時 **依序拼接**，段落標題可不寫入 string，但 **語義順序不可調換**。

### A. HARD OUTPUT

- Exact **320×400** canvas  
- **One subject** only  
- Solid flat background **`#1A1410`**  
- **No** sprite sheet, collage, decorative frame, border, caption, watermark, UI, text

### B. STYLE LOCK

- **Cartoon pixel art**, crisp integer pixels  
- **1px hard black outline** on character silhouette  
- **No anti-aliasing**, no soft blur, no photoreal  
- **Top-left light** (shadow falls down-right)  
- Palette: **base neutrals on `#1A1410`** + **faction accent colors**（見 D 陣營色律）

### C. POSE LOCK

- **Front-facing half-body** portrait  
- **Head : body ≈ 1 : 2.3**（同 batch shared skeleton）  
- Shoulders level; arms hang naturally; **open hands**（非握拳，除非 IDENTITY 明確例外）  
- Same stance across all batch1 character cards（肩線、鏡頭距離、手位置對齊）

### D. IDENTITY SLOT（per character · 必填）

以 **「MUST SHOW FIRST」** 列出可審核 landmark，順序固定：

1. **Hat / headgear**（帽、冠、巾）  
2. **Face**（鬚、膚色、表情）  
3. **Costume color**（主色 +  mandatory 配件色，如 teal sash）  
4. **Weapon OR prop**（或 explicit **no polearm**）  
5. **Faction color law**（魏／蜀／吳點綴與 **禁止** 色，如周瑜禁大面積綠）

可附 **approximate canvas %** 作 art-manager 量尺（非 API 必須）。

### E. FORBIDDEN（always append · 必 appended）

全局 + 角色專屬禁止項，**永遠放在 string 末尾**，以 `Forbidden:` 起首。

**Global（每卡必含）**：second person, horse, mount, full background scene, photoreal, gradient or textured background, wrong size not 320x400, sprite sheet, collage, frame, text, logo, missing hands, clenched fists when POSE requires open hands, anti-aliased soft edges

---

## 4. Empty fillable template · 空白可填模板

複製下方_fence 內容，將 `{{...}}` 替換後 **刪除所有 `{{` `}}`**，合成 **一行或一段** 提交 API。

```
A. HARD OUTPUT: exact 320x400 pixel character card, single subject only, solid flat background #1A1410, no sprite sheet no collage no decorative frame no border no text no watermark,
B. STYLE LOCK: cartoon pixel art, 1px hard black outline, no anti-aliasing no photoreal, top-left light shadow down-right, palette {{BASE_NEUTRALS}} with accents {{ACCENT_COLORS}},
C. POSE LOCK: front half-body portrait facing viewer, head to body ratio 1:2.3, shoulders level arms hanging naturally open hands not fists, shared batch1 skeleton same shoulder height and camera distance,
D. IDENTITY {{CHARACTER_ID}} MUST SHOW FIRST: (1) hat/headgear {{HAT_LANDMARKS}}, (2) face {{FACE_LANDMARKS}}, (3) costume color {{COSTUME_COLORS}}, (4) weapon or prop {{WEAPON_OR_PROP}}, (5) faction color law {{FACTION_COLOR_LAW}},
E. FORBIDDEN: {{GLOBAL_FORBIDDEN}} {{CHARACTER_FORBIDDEN}}
```

---

## 5. Filled examples · 填寫範例（**非新版本**）

以下三條為 **STANDARD v1 下的範例**，細節密度對齊歷史 v06 ultra，**標籤為 Example only** — 後續改動請 **只改 STANDARD 或增補範例段落**，不新增 `p0_chars_v0N_*` 檔。

### Example A — 曹操 Cao Cao · `PX2D_CHAR_WEI_Caocao_ex`

```
A. HARD OUTPUT: exact 320x400 pixel character card, single subject only, solid flat background #1A1410, no sprite sheet no collage no decorative frame no border no text no watermark,
B. STYLE LOCK: cartoon pixel art, 1px hard black outline, no anti-aliasing no photoreal, top-left light shadow down-right, palette dark teal black and gold accents on #1A1410,
C. POSE LOCK: front half-body portrait facing viewer, head to body ratio 1:2.3, shoulders level y44-48 percent arms hanging naturally open hands resting near sash not fists, shared batch1 skeleton,
D. IDENTITY Cao Cao Wei MUST SHOW FIRST: (1) hat black winged guan two tall upright wing fins wing tips y4-8 gold brow band y18-22 not plain helmet, (2) face thin goatee tip y34-38 mature calm eyes no full beard, (3) costume dark teal Wei robe mandatory bright teal waist sash y52-58 at least eight percent teal read on torso, (4) weapon or prop empty hands NO spear NO halberd NO polearm, (5) faction color law Wei teal plus black plus gold forbid brown-gold-only without visible teal,
E. FORBIDDEN: second person horse mount full scene photoreal gradient background wrong size not 320x400 sprite sheet collage frame text anti-aliased soft edges plain round helmet no wings brown and gold only costume without teal spear halberd any polearm
```

### Example B — 張飛 Zhang Fei · `PX2D_CHAR_SHU_Zhangfei_ex`

```
A. HARD OUTPUT: exact 320x400 pixel character card, single subject only, solid flat background #1A1410, no sprite sheet no collage no decorative frame no border no text no watermark,
B. STYLE LOCK: cartoon pixel art, 1px hard black outline, no anti-aliasing no photoreal, top-left light shadow down-right, palette black brown tiger fur with Shu green accent trim only,
C. POSE LOCK: front half-body portrait facing viewer, head to body ratio 1:2.3, shoulders level y44-48 percent arms hanging both hands open palms y68-78 NOT fists, shared batch1 skeleton,
D. IDENTITY Zhang Fei Shu MUST SHOW FIRST: (1) hat simple battle hair tie not dominating beard, (2) face darker fierce complexion fierce eyes y24-28 massive black beard covering chest y28-54 width x28-72, (3) costume symmetric tiger fur pelt both shoulders x12-30 and x70-88 y40-52 Shu green accent under collar only not dominant, (4) weapon upright serpentine spear snake mao snaky curved blade tip visible y8-18 shaft x72-88, (5) faction color law Shu green accents OK under pelt forbid pale scholar face and small tidy beard,
E. FORBIDDEN: second person horse mount full scene photoreal gradient background wrong size not 320x400 sprite sheet collage frame text anti-aliased soft edges small beard single shoulder fur hidden spear tip clenched fists gentle pale face
```

### Example C — 周瑜 Zhou Yu · `PX2D_CHAR_WU_Zhouyu_ex`

```
A. HARD OUTPUT: exact 320x400 pixel character card, single subject only, solid flat background #1A1410, no sprite sheet no collage no decorative frame no border no text no watermark,
B. STYLE LOCK: cartoon pixel art, 1px hard black outline, no anti-aliasing no photoreal, top-left light shadow down-right, palette red gold black scholar armor on #1A1410,
C. POSE LOCK: front half-body portrait facing viewer, head to body ratio 1:2.3, shoulders level y44-48 percent arms hanging open hands optional fan one hand, shared batch1 skeleton,
D. IDENTITY Zhou Yu Wu MUST SHOW FIRST: (1) hat elegant tall black scholar guan crown top y5-9 height ten-fourteen percent NOT feather helm NOT Wei winged crown, (2) face young handsome y20-34 calm expression, (3) costume long red silk head neck scarf knot y18-24 trailing ends y45-70 at least twenty-five percent height one side red gold armor chest y40-72, (4) weapon or prop optional folding fan x55-75 y58-68 no long polearm, (5) faction color law Wu red gold dominant forbid large green fabric over ten percent green area,
E. FORBIDDEN: second person horse mount full scene photoreal gradient background wrong size not 320x400 sprite sheet collage frame text anti-aliased soft edges large green cloak dominant green costume Wei winged guan feather warrior helmet
```

---

## 6. Governance · 修訂規則

| 允許 | 禁止 |
|------|------|
| 修訂 **`PROMPT_STANDARD_char_card_v1.md`**（版本 bump 僅在檔名 `v1→v2` 時） | 新增 `p0_chars_v07_*`、`v08_detail` 等平行規格檔 |
| 在本檔 **§5 增刪 Example** 或 Notion 上登記新 IDENTITY 附錄 | 以「 ultra / detail 」檔取代 STANDARD 作為主交付 |
| 歷史 v04/v05/v06 保留只讀對照 | 在歷史檔頂部寫回「canonical」 |

---

## 7. Related · 關聯

- 批次 WIP：[`p0_batch1_chars_notes.md`](./p0_batch1_chars_notes.md)  
- 戰場單位：**[`PROMPT_STANDARD_unit_sprite_v1.md`](./PROMPT_STANDARD_unit_sprite_v1.md)**（128×128）  
- API 紀錄：[`poc_v01_notes.md`](./poc_v01_notes.md)
