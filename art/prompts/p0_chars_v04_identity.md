# P0 角色卡 — v04 identity（canonical）

> **v04** 為角色 **identity + 調色** 的單一真相來源（可取代或與 `p0_chars_v03e_landmark.md` 並用：v03e 保留 landmark／尺寸閘細節；**生成時以 v04 全文 positive 為準**）。**PR #20** pose 鎖仍有效；**landmark + 精確 320×400** 為強制閘。

## Hard size gate

- Output file MUST be exactly **320×400** px (`identify` / PIL after gen; wrong → Nearest-neighbor crop/resize to 320×400; unusable → re-gen).
- Single subject; solid `#1A1410`; no collage / 九宮格 / gold frame / text / UI chrome.
- Unified front half-body pose (same skeleton across 3 cards).

## Shared style

Cartoon pixel (chibi + hard 1px outline), left-top light, head:body ~1:2.2–1:2.5, shared color ramp.

## Identity color rules (mandatory)

| 角色 | 調色閘 |
|------|--------|
| 曹操 Wei | **Teal sash / teal Wei trim mandatory** — robe may have dark armor; **NOT brown-gold-only** palette (no missing teal). |
| 張飛 Shu | **Tiger-fur shoulder pelt + upright serpent spear (蛇矛) + dark swarthy face** under beard; all three visible. |
| 吳將 Zhou Yu 風 | **Red-and-gold Wu armor + long silk neck scarf**; **NO large green fabric** (no green cape/cloak/dominant green panels). |

## Copy-paste prompts (single positive block; negatives inline)

Each block is one paste — landmark first (~top 30% of frame), size gate, identity colors, and **avoid:** clauses at end.

### 曹操 / Cao Cao (Wei)

```
MUST SHOW FIRST: black winged guan crown with two tall upright wing fins filling the top of the frame, stern middle-aged Three Kingdoms warlord Cao Cao, teal-dark Wei robe with visible mandatory teal sash and teal trim NOT brown-gold-only costume, dark armor accents, cartoon pixel art chibi half-body front view, hard 1px black outline, flat color, left-top lighting, solid #1A1410 background, single character only, exact 320x400 game character card output, avoid: plain round helmet, generic warrior helm, missing winged guan, missing teal sash, brown-gold-only robe without teal, modern face, soft anti-aliased edges, collage, grid, nine-panel, gold ornate frame, text, watermark, multiple characters, photorealistic
```

### 張飛 / Zhang Fei (Shu)

```
MUST SHOW FIRST: massive black beard covering the entire chest, fierce Three Kingdoms general Zhang Fei with dark swarthy face visible around the beard, tiger-fur shoulder pelt on both shoulders, huge serpent spear snake-mao held upright beside the body, Shu green-brown armor accents, cartoon pixel art chibi half-body front view, hard 1px black outline, flat color, left-top lighting, solid #1A1410 background, single character only, exact 320x400 game character card output, avoid: clean-shaven, pale face, short beard, missing tiger pelt, missing spear, plain helmet, collage, gold frame, text, soft blurry pixels, multiple characters, photorealistic
```

### 吳將／周瑜風 (Wu placeholder)

```
MUST SHOW FIRST: long flowing silk neck scarf around the neck, elegant guan hat, handsome Three Kingdoms strategist like Zhou Yu, red-and-gold Wu armor panels, cartoon pixel art chibi half-body front view, hard 1px black outline, flat color, left-top lighting, solid #1A1410 background, single character only, exact 320x400 game character card output, avoid: missing scarf, plain helmet, generic soldier, large green fabric, green cape, green cloak, dominant green costume panels, collage, gold frame, text, soft edges, multiple characters, photorealistic
```
