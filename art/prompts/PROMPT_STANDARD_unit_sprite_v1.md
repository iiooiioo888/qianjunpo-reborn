# PROMPT STANDARD — 戰場單位 sprite AIGC v1（**canonical · locked**）

> 128×128 戰場單位／idle sprite。**結構對齊** [`PROMPT_STANDARD_char_card_v1.md`](./PROMPT_STANDARD_char_card_v1.md)（A→E 順序）；僅 **硬輸出尺寸** 與 **POSE** 改為全身高小圖。  
> **禁止** 新增 `p0_units_v03_*` 等平行規格檔作為主交付；修訂本檔或增補 Example。  
> [`p0_units_v02.md`](./p0_units_v02.md) 為 **historical** 路徑／調色備忘。

---

## 1. Purpose · 目的

wan2.7-image 生成 **128×128** 卡通像素單位（如步兵、騎兵），與角色卡 **同一 STYLE LOCK** 與 **`#1A1410`** 背景，供 `PX2D_unit_*` WIP 審核。

---

## 2. Pipeline · 流程

填模板 → 單条 positive（E 段 Forbidden 內嵌）→ API → 驗證 **128×128** → Notion → art-manager。

---

## 3. Sections A→E（unit 專用差異）

| 段 | Unit 規則 |
|----|-----------|
| **A. HARD OUTPUT** | Exact **128×128**, one subject, **`#1A1410`**, no sheet/collage/frame/text |
| **B. STYLE LOCK** | Cartoon pixel, **1px outline**, no AA, top-left light；調色可沿用 **`#8B4513` / `#CD5C5C` / `#DAA520`** 作兵種基色 + 陣營小面積點綴 |
| **C. POSE LOCK** | **Front or 3/4 front idle**（同 batch 固定一種）；全身高在畫幅內 **85–95%** 高度；**open hands** 持械自然；與 [`PROMPT_STANDARD_char_card_v1.md`](./PROMPT_STANDARD_char_card_v1.md) **同一 idle 骨架語言** |
| **D. IDENTITY SLOT** | MUST SHOW FIRST：兵種輪廓（步兵盾矛／騎兵馬上或步態）→ 盔/髮 → 主色塊 → 武器 → 陣營色律 |
| **E. FORBIDDEN** | 同角色卡 global + `wrong size not 128x128` + 多角色 + 馬（除非 D 明確為 cavalry 單位且 **one mount one rider still one subject** 依 art-manager 定義） |

---

## 4. Empty template · 空白模板

```
A. HARD OUTPUT: exact 128x128 pixel unit sprite, single subject only, solid flat background #1A1410, no sprite sheet no collage no frame no text,
B. STYLE LOCK: cartoon pixel art, 1px hard black outline, no anti-aliasing, top-left light, palette {{UNIT_BASE_COLORS}} accents {{FACTION_ACCENTS}},
C. POSE LOCK: {{IDLE_POSE}} full figure height 85-95 percent of canvas, shared batch idle skeleton with character cards,
D. IDENTITY {{UNIT_ID}} MUST SHOW FIRST: (1) silhouette {{SILHOUETTE}}, (2) headgear {{HEAD}}, (3) armor/cloth color {{COLORS}}, (4) weapon {{WEAPON}}, (5) faction law {{FACTION_LAW}},
E. FORBIDDEN: second character duplicate unit wrong size not 128x128 photoreal gradient background sprite sheet collage text frame {{UNIT_FORBIDDEN}}
```

---

## 5. Examples · 範例（非新版本）

### Example — 步兵 Infantry · `PX2D_unit_infantry`

```
A. HARD OUTPUT: exact 128x128 pixel unit sprite, single subject only, solid flat background #1A1410, no sprite sheet no collage no frame no text,
B. STYLE LOCK: cartoon pixel art, 1px hard black outline, no anti-aliasing, top-left light, palette #8B4513 leather #CD5C5C cloth trim #DAA520 metal accents,
C. POSE LOCK: front idle standing full body height 90 percent canvas feet near bottom open hands holding spear vertically at side shared batch idle skeleton,
D. IDENTITY infantry MUST SHOW FIRST: (1) silhouette foot soldier with round shield on left arm, (2) simple helmet or head wrap, (3) brown-red uniform blocks readable at 128px, (4) short spear and wooden shield, (5) faction accent small colored sash only not dominant,
E. FORBIDDEN: wrong size not 128x128 second soldier horse photoreal gradient background sprite sheet collage text frame cavalry mount
```

### Example — 騎兵 Cavalry · `PX2D_unit_cavalry`

```
A. HARD OUTPUT: exact 128x128 pixel unit sprite, single subject only, solid flat background #1A1410, no sprite sheet no collage no frame no text,
B. STYLE LOCK: cartoon pixel art, 1px hard black outline, no anti-aliasing, top-left light, palette #8B4513 #CD5C5C #DAA520 with dark horse coat,
C. POSE LOCK: three-quarter front idle rider plus horse single composed subject full group height 92 percent canvas shared batch idle skeleton,
D. IDENTITY cavalry MUST SHOW FIRST: (1) silhouette horse and rider one icon readable, (2) rider helmet plume small, (3) horse brown dark mane rider red-brown armor, (4) lance angled backward not covering face, (5) faction accent on rider sash only,
E. FORBIDDEN: wrong size not 128x128 infantry foot-only duplicate riders photoreal gradient background sprite sheet collage text frame
```

---

## 6. Historical

- [`p0_units_v02.md`](./p0_units_v02.md) — v02 路徑與調色；**生成依據以本 STANDARD 為準**。
