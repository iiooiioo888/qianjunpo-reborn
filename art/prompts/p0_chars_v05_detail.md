# P0 batch1 — 角色卡 AIGC 提示詞 v05（detail）

> **Historical only**：**不得** 作為生成依據。Canonical：**[`PROMPT_STANDARD_char_card_v1.md`](./PROMPT_STANDARD_char_card_v1.md)**（曹操／張飛／周瑜填寫範例在 STANDARD §5，非新版本）。  
> 歷史：v05 曾取代 [v04](./p0_chars_v04.md)；後有 [v06 ultra](./p0_chars_v06_ultra.md) — 均已 archive。

## 硬門檻（Hard gate — 全角色）

- **320×400** exact canvas, **single subject**, flat background **`#1A1410`**
- **Cartoon pixel art**, crisp pixels, **1px black outline** on all silhouettes
- **Shared front half-body pose**: facing camera, shoulders level, both arms visible to mid-forearm, identical stance across batch1 (no dynamic action pose)
- **Forbidden (global)**: second character, horse, full background scene, photoreal, smooth gradient, missing hands, wrong aspect ratio, soft blur anti-aliasing

---

## 曹操 · CAO CAO (`PX2D_CHAR_WEI_Caocao_ex`)

**MUST SHOW FIRST（審圖順序）**

1. **Headgear**: black **winged guan** — two **tall upright wing fins** flanking the crown (not a plain round helmet); **gold brow band** across forehead
2. **Face**: **thin goatee** (小髭山羊鬚), mature calm eyes, no full beard
3. **Costume**: **dark teal** Wei robe; **mandatory teal sash** at waist (must read as teal, not brown belt)
4. **Weapon**: **no spear**, no halberd, empty hands or subtle hand on sash only
5. **Color**: teal + black + gold accents; **forbid brown-gold-only** palette (must include visible teal)

### Copy-paste positive block (v05)

```
320x400 cartoon pixel art character card, single subject only, flat solid background #1A1410, 1px black pixel outline, shared front half-body pose facing viewer, Cao Cao Wei warlord, FIRST read black winged guan with two tall upright wing fins plus gold brow band across forehead, thin neat goatee no full beard, dark teal Wei robe with mandatory bright teal waist sash clearly visible, hands empty or resting near sash NO spear NO halberd, mature dignified face, teal black and gold accents, crisp pixels no anti-aliasing, Forbidden: plain round helmet, single crest only, brown and gold only costume without teal, spear weapon, second person, horse, photoreal, gradient background, wrong size not 320x400
```

---

## 張飛 · ZHANG FEI (`PX2D_CHAR_SHU_Zhangfei_ex`)

**MUST SHOW FIRST**

1. **Beard**: **massive black beard** covering upper chest (dominant silhouette)
2. **Shoulders**: **tiger-fur pelt** on **both shoulders** (symmetric pelts)
3. **Weapon**: **upright serpentine spear 蛇矛** with **curved snaky blade tip visible** above shoulder line
4. **Face**: **darker fierce** complexion, wide fierce eyes under brow
5. **Hands**: **open hands** relaxed at sides — **not fists**
6. **Color**: Shu **green accents OK** under / at edges of pelt (collar trim, inner robe)

### Copy-paste positive block (v05)

```
320x400 cartoon pixel art character card, single subject only, flat solid background #1A1410, 1px black pixel outline, shared front half-body pose facing viewer, Zhang Fei Shu general, FIRST read massive chest-covering black beard, tiger fur pelt on both shoulders symmetric, upright serpentine-blade spear 蛇矛 with snaky curved tip visible, darker fierce face, Shu green accent trim on robe under pelt OK, both hands open and relaxed at sides NOT clenched fists, crisp pixels no anti-aliasing, Forbidden: small tidy beard, single shoulder fur only, hidden spear, pale gentle face, fist hands, second person, horse, photoreal, gradient background, wrong size not 320x400
```

---

## 周瑜 · ZHOU YU / 吳 (`PX2D_CHAR_WU_Zhouyu_ex`)

**MUST SHOW FIRST**

1. **Scarf**: **long red silk head and neck scarf** with **trailing ends** (must hang past shoulder)
2. **Headgear**: **elegant tall black scholar guan** (文士高冠) — **not** feather helm, **not** Wei winged crown
3. **Armor**: **red-gold** lamellar or scale armor chest
4. **Face**: **young handsome** scholar-general
5. **Prop**: optional **folding fan** in one hand
6. **Forbidden**: **large green fabric** panels, cloaks, or dominant green costume

### Copy-paste positive block (v05)

```
320x400 cartoon pixel art character card, single subject only, flat solid background #1A1410, 1px black pixel outline, shared front half-body pose facing viewer, Zhou Yu Wu commander young handsome, FIRST read long red silk head neck scarf with trailing scarf ends past shoulders, elegant tall black scholar guan NOT feather helm NOT Wei winged crown, red and gold armor on chest, optional folding fan, crisp pixels no anti-aliasing, Forbidden: large green fabric cloak or dominant green costume, Wei winged guan, feathered warrior helm, second person, horse, photoreal, gradient background, wrong size not 320x400
```

---

## 交叉引用

- 批次說明：[`p0_batch1_chars_notes.md`](./p0_batch1_chars_notes.md)
- **生成請只用**：[`PROMPT_STANDARD_char_card_v1.md`](./PROMPT_STANDARD_char_card_v1.md)
