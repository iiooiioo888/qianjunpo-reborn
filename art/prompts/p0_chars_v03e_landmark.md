# P0 角色卡 — v03e landmark + 320×400 尺寸閘

> **v03e** 取代 v02 薄筆記中關於 landmark 與尺寸的說明；**PR #20** 的 pose 鎖定筆記仍適用於統一骨架，但 **landmark 佔位 + 精確 320×400** 為強制閘門（不通過即重製／裁切修正）。

## Hard size gate

- Character card output file MUST be exactly **320×400** pixels (verify with `identify` or PIL after gen; if wrong, Nearest-neighbor crop/resize to 320×400; if unusable, re-gen).
- Single subject only; solid bg `#1A1410`; no collage / 九宮格 / gold frame / text / UI chrome.
- Unified front half-body pose (same skeleton across 3 cards).

## Shared style

Cartoon pixel (chibi + hard 1px outline), left-top light, fixed head:body ~1:2.2–1:2.5, same color ramp.

## Per-character prompts (English positive MUST open with landmark occupying top ~30% of frame)

### 曹操 / Cao Cao (Wei)

POSITIVE (copy-paste ready):

```
MUST SHOW FIRST: black winged guan crown with two tall upright wing fins filling the top of the frame, teal-dark Wei robe with dark armor accents, stern middle-aged Three Kingdoms warlord, cartoon pixel art chibi half-body front view, hard 1px black outline, flat color, left-top lighting, solid #1A1410 background, single character only, 320x400 game character card
```

NEGATIVE:

```
plain round helmet, generic warrior helm, missing winged guan, modern face, soft anti-aliased edges, collage, grid, nine-panel, gold ornate frame, text, watermark, multiple characters, photorealistic
```

### 張飛 / Zhang Fei (Shu)

POSITIVE:

```
MUST SHOW FIRST: massive black beard covering the entire chest, tiger-fur shoulder pelt, huge serpent spear (蛇矛) held upright beside the body, fierce Three Kingdoms general Zhang Fei, cartoon pixel art chibi half-body front view, hard 1px black outline, flat color, left-top lighting, solid #1A1410 background, single character only, 320x400 game character card
```

NEGATIVE:

```
clean-shaven, short beard, missing tiger pelt, missing spear, plain helmet, collage, gold frame, text, soft blurry pixels, multiple characters
```

### 吳將／周瑜風 (Wu placeholder)

POSITIVE:

```
MUST SHOW FIRST: long flowing silk neck scarf, elegant guan hat, red-and-gold Wu armor, handsome Three Kingdoms strategist like Zhou Yu, cartoon pixel art chibi half-body front view, hard 1px black outline, flat color, left-top lighting, solid #1A1410 background, single character only, 320x400 game character card
```

NEGATIVE:

```
missing scarf, plain helmet, generic soldier, collage, gold frame, text, soft edges, multiple characters
```
