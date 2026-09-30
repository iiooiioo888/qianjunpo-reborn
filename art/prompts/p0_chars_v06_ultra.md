# P0 batch1 — 角色卡 AIGC 提示詞 v06（ultra · **canonical**）

> **Canonical for generation**：ComfyUI／worker／人工重跑 **僅引用本檔**。  
> Supersedes：[`p0_chars_v05_detail.md`](./p0_chars_v05_detail.md) · [`p0_chars_v04.md`](./p0_chars_v04.md)（兩檔頂部已標 **superseded by v06**）。

## 座標系（320×400）

- 原點：左上 `(0,0)`；**x%** = `x/320`，**y%** = `y/400`（以下為 **approximate** 審圖錨點，±3% 容忍）。
- **Shared pose anchors（三人相同）**

| Anchor | y% | 說明 |
|--------|-----|------|
| 頭頂髮/冠頂 | **6–10%** | 最高像素不得觸及 y=0 以外留白 >4px |
| 下巴 | **36–40%** | 下頜最低點 |
| 肩線 | **44–48%** | 左右肩同高 |
| 肘部 | **58–64%** | 手臂自然下垂 |
| 畫幅底裁切 | **92–98%** | 半身止於上腹/腰帶，不切大腿 |

| Anchor | x% | 說明 |
|--------|-----|------|
| 身體中軸 | **50%** | 鼻子/ sternum 對齊 |
| 肩寬外緣 | **22% / 78%** | 不含武器超出時 |

**Hard gate（全角色）**

- **320×400 exact**, **single subject**, background **`#1A1410` only**
- **Cartoon pixel**, **1px black outline**, no AA blur
- **Shared front half-body pose**（上表 anchors 一致）
- **Global Forbidden**: second character, mount, full scene, text/watermark, photoreal, soft gradient bg, wrong dimensions, cropped missing hands, duplicate faces

**資產 ID**

- `PX2D_CHAR_WEI_Caocao_ex` — 曹操
- `PX2D_CHAR_SHU_Zhangfei_ex` — 張飛
- `PX2D_CHAR_WU_Zhouyu_ex` — 周瑜

---

## 曹操 · CAO CAO

### BLOCK — HAT

| 元素 | 規則 | Canvas % |
|------|------|----------|
| Winged guan 翼冠 | 兩片 **豎直翼片**（wing fins）在冠兩側，高於冠頂 | 翼片 tip **y 4–8%**；翼高 **12–18%** 畫幅 |
| Gold brow band | 額前 **金帶** 橫貫 | **y 18–22%**，寬 **x 38–62%** |
| 禁止 | plain helmet、無翼圓盔、魏翼冠用在周瑜 | — |

### BLOCK — FACE

| 元素 | 規則 | Canvas % |
|------|------|----------|
| Goatee | **細長山羊鬚**，無絡腮 | 鬚尖 **y 34–38%** |
| 膚色 | 偏黃的成熟肤色 | 頰部 **x 35–65%**, **y 22–34%** |
| 眉眼神 | 沉穩、略挑眉 | 眼線 **y 24–28%** |

### BLOCK — BODY

| 元素 | 規則 | Canvas % |
|------|------|----------|
| Wei robe | **深青綠 dark teal** 長袍 | 軀幹 **y 38–88%** |
| Teal sash | **必須可讀的 teal 腰帶**（非純褐） | **y 52–58%**，寬 **6–10%** 畫高 |
| 手 | 空手或指節搭在腰帶 | 腕 **y 62–72%** |

### BLOCK — WEAPON

| 元素 | 規則 |
|------|------|
| 禁止長柄 | **NO spear**, NO halberd, NO polearm of any kind |

### BLOCK — COLOR

| 必須 | 禁止 |
|------|------|
| **teal + black + gold** 同屏可辨 | **brown-gold-only**（無 teal 主色） |
| 腰帶/teal 面積 ≥ **8%** 軀幹 | 全褐魏袍冒充 teal |

### Single-string positive（copy-paste · Forbidden embedded）

```
320x400 cartoon pixel art character card exact size, single subject Cao Cao only, flat solid background #1A1410, 1px black pixel outline crisp no anti-aliasing, shared front half-body pose anchors head top y6-10 chin y36-40 shoulders y44-48, FIRST silhouette black winged guan two tall upright wing fins wing tips y4-8 fin height 12-18 percent canvas gold brow band y18-22, thin goatee tip y34-38 no full beard, dark teal Wei robe torso y38-88 mandatory bright teal waist sash y52-58 at least eight percent teal read on torso, empty hands no spear no halberd no polearm, teal black gold palette, Forbidden: plain round helmet no wings, brown and gold only costume without visible teal, spear halberd weapon, second person horse, photoreal gradient background wrong aspect not 320x400, Wei wings on wrong character
```

---

## 張飛 · ZHANG FEI

### BLOCK — HAT / HEAD

| 元素 | 規則 | Canvas % |
|------|------|----------|
| 髮髻/巾 | 簡素戰髮或短巾，**不搶鬍** | 冠飾 **y 8–16%** |
| 眉 | 粗眉壓眼 | **y 22–26%** |

### BLOCK — FACE

| 元素 | 規則 | Canvas % |
|------|------|----------|
| 膚色 | **偏深、兇悍** | 頰 **y 24–32%** |
| 眼 | 圓睜、怒目 | 眼 **y 24–28%** |
| 鬍 | **massive black beard** 覆蓋胸上部 | 鬍上緣 **y 28%** 下緣 **y 48–54%**，寬 **x 28–72%** |

### BLOCK — BODY

| 元素 | 規則 | Canvas % |
|------|------|----------|
| Tiger pelt | **雙肩**虎裘對稱 | 左肩 **x 12–30%** 右肩 **x 70–88%**, **y 40–52%** |
| 內袍 | 蜀 **green accent** 可在裘下緣/領口 | 綠面積 **≤15%** 軀幹，非主色 |
| 手 | **張開**自然下垂 | 掌 **y 68–78%**, **open fingers not fists** |

### BLOCK — WEAPON

| 元素 | 規則 | Canvas % |
|------|------|----------|
| 蛇矛 | 豎立於身側，**蛇形刃尖**必見 | 矛尖 **y 8–18%**；杆 **x 72–88%** |

### BLOCK — COLOR

| 必須 | 禁止 |
|------|------|
| 黑鬍 + 虎裘棕黑 + 矛杆深木色 | 小鬍、白面書生臉 |
| 蜀綠僅點綴 | 綠色主導整袍（留給周瑜禁令不同） |

### Single-string positive（copy-paste · Forbidden embedded）

```
320x400 cartoon pixel art character card exact size, single subject Zhang Fei only, flat solid background #1A1410, 1px black pixel outline crisp no anti-aliasing, shared front half-body pose anchors head top y6-10 chin buried in beard shoulders y44-48, FIRST read massive black beard covering chest y28-54 width x28-72, symmetric tiger fur pelt both shoulders x12-30 and x70-88 y40-52, darker fierce face y24-32, upright serpentine spear 蛇矛 beside body shaft x72-88 snaky curved blade tip visible y8-18, Shu green accent trim under pelt collar only not dominant, both hands open relaxed at sides y68-78 open palms NOT fists, Forbidden: small tidy beard, single shoulder fur, hidden spear tip, pale gentle scholar face, clenched fists, second person horse, photoreal gradient background wrong aspect not 320x400
```

---

## 周瑜 · ZHOU YU（吳）

### BLOCK — HAT

| 元素 | 規則 | Canvas % |
|------|------|----------|
| Scholar guan | **高瘦黑文士冠**，方/梯形轮廓 | 冠頂 **y 5–9%**，冠高 **10–14%** |
| 禁止 | feather helm、Wei **winged** crown | — |

### BLOCK — FACE

| 元素 | 規則 | Canvas % |
|------|------|----------|
| 年齡 | **young handsome** | 膚 **y 20–34%** |
| 表情 | 從容微笑或沉靜 | 眼 **y 24–28%** |

### BLOCK — BODY（含 SCARF）

| 元素 | 規則 | Canvas % |
|------|------|----------|
| Red silk scarf | 包頭至頸，**飄帶末端** | 結 **y 18–24%**；尾端 **y 45–70%**, **x 15–35% 或 65–85%** |
| Scarf length | 至少一側飄帶 **≥25%** 畫高 | 見上 |
| Red-gold armor | 胸甲 **红金** 鱗/札 | **y 40–72%**, **x 30–70%** |
| Fan（可選） | 折扇於一手 | 扇 **x 55–75%**, **y 58–68%** |

### BLOCK — WEAPON

| 元素 | 規則 |
|------|------|
| 默認 | 無長柄兵器；可持扇 |

### BLOCK — COLOR

| 必須 | 禁止 |
|------|------|
| **red scarf + red-gold armor + black guan** | **large green fabric**（大綠披風/大綠袍） |
| 綠 | 僅允許 **≤5%** 細線鑲邊若必須 | 任何 **>10%** 綠色面積 |

### Single-string positive（copy-paste · Forbidden embedded）

```
320x400 cartoon pixel art character card exact size, single subject Zhou Yu only, flat solid background #1A1410, 1px black pixel outline crisp no anti-aliasing, shared front half-body pose anchors head top y6-10 chin y36-40 shoulders y44-48, FIRST read long red silk head neck scarf knot y18-24 trailing ends y45-70 at least twenty-five percent canvas height on one side, elegant tall black scholar guan crown top y5-9 height ten-fourteen percent NOT feather helm NOT Wei winged crown, young handsome face y20-34, red and gold armor chest y40-72 x30-70, optional folding fan x55-75 y58-68, no long polearm, Forbidden: large green fabric cloak or dominant green costume over ten percent green area, Wei winged guan on Zhou Yu, feathered warrior helmet, second person horse, photoreal gradient background wrong aspect not 320x400
```

---

## 生成檢查清單（art-manager）

1. 像素尺 **320×400** 與 **`#1A1410`** 背景取样
2. 三人 **肩 y44-48%** 對齊（同 batch pose）
3. 各角色 **FIRST read** 是否命中對應 BLOCK 首項
4. **Forbidden** 項：曹操無矛／周瑜無大綠／張飛非握拳

## 交叉引用

- v04 / v05：僅存檔與 diff 對照，**勿再生成**
- WIP 輸出：`art/2d/_wip/characters/PX2D_CHAR_*`（見 [`p0_batch1_chars_notes.md`](./p0_batch1_chars_notes.md)）
