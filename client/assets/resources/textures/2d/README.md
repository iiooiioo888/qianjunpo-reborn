# Cocos `resources/textures/2d`

戰術客戶端 2D 像素貼圖匯入位（**非** `art/` 正式管線）。顯示層一律經 **Registry 邏輯鍵** 解析路徑；換圖只改 registry／檔名，不動玩法腳本。

## 單位（`textures/2d/units/`）

| 邏輯鍵 (`UnitTextureKey`) | 預設 stem（無擴展名） | `ViewUnit.type` |
|---------------------------|----------------------|-----------------|
| `infantry` | `PX2D_unit_infantry` | `0`（預設） |
| `cavalry` | `PX2D_unit_cavalry` | `2` |

- **STANDARD 畫布**：128×128（常數 `UNIT_ART_CANVAS_PX`）；棋盤格內僅 **Nearest** + **整數倍** 縮放（1×／2× 或反覆 ÷2），見 `PixelSpriteUtil.boardUnitDisplaySize`。
- **換 stem**：將 PNG 放入本目錄 `units/`，檔名與 stem 一致（或改 `UnitSpriteRegistry` 內 stem／呼叫 `registerResourcePath('infantry', '…')` 後 `preload()`）。
- **同步 STANDARD 工作稿**（來源 art #48，不改 `art/`）：

```bash
bash client/scripts/sync-wip-unit-textures.sh
```

載入：`UnitSpriteRegistry.getSpriteFrameByKey('infantry')` 或 `applyUnitToSprite`；缺圖時 `UnitPlaceholderView` 幾何占位。

## 角色卡（`textures/2d/chars/`）

| 邏輯鍵 (`CharCardTextureKey`) | 預設 stem |
|-------------------------------|-----------|
| `char_caocao` | `PX2D_CHAR_WEI_Caocao_ex_v04` |
| `char_zhangfei` | `PX2D_CHAR_SHU_Zhangfei_ex_v04` |
| `char_wu_placeholder` | `PX2D_CHAR_WU_Placeholder_01_v04` |

- **v04 畫布**（P0 identity）：320×400（`CHAR_CARD_ART_WIDTH_PX` × `CHAR_CARD_ART_HEIGHT_PX`）；HUD 預覽用整數倍縮入框，見 `charCardDisplaySize`。
- **缺 PNG 時**：`CharacterCardHudStrip` 顯示色塊占位，不 crash。
- **同步 v04 工作稿**（不改 `art/`）：
- **換貼集中入口**：`assets/scripts/display/TextureRegistryDev.ts` 的 `applyDevTextureStemOverrides()` → `preloadTacticalDisplayTextures()`（`TacticalBootstrap` 啟動時已呼叫）。
- **可選 client 極簡占位**（非 art、勿從 `_wip` 複製）：例如 `units/mock_infantry_128.png`（128×128）、`chars/mock_char_caocao_320x400.png`（320×400），再在 `TextureRegistryDev` 內 `registerResourcePath` 指向 stem。

```bash
bash client/scripts/sync-wip-char-textures.sh
```

## 地格（`textures/2d/tiles/`）

| 邏輯鍵 (`TerrainTileTextureKey`) | 預設 stem | `TerrainKind` |
|----------------------------------|-----------|---------------|
| `plain` | `ISO25_tile_grass_v02` | `Plain`（`0`） |
| `mountain` | `ISO25_tile_mountain_v01` | `Mountain`（`1`） |

- **STANDARD 畫布**：64×32 等距菱形（`ISO25_TILE_ART_WIDTH_PX` × `ISO25_TILE_ART_HEIGHT_PX`）；棋盤 **Nearest** + **整數倍** 縮放，見 `boardIsoTileDisplaySize`。
- **同步 STANDARD 工作稿**（來源 art #83，不改 `art/`）：

```bash
bash client/scripts/sync-wip-tile-textures.sh
```

載入：`TerrainTileSpriteRegistry.getSpriteFrameForTerrain(terrain)`；缺圖時 `TacticalBoardView` 平面色塊占位。

## 圖標（`textures/2d/icons/`）

| 檔案 | 用途（`IconAssetId`） |
|------|------------------------|
| `icons/PX2D_icon_res_*_32.png` 等 | HUD `ResourceIconHudStrip` |
| `icons/PX2D_icon_bld_*_32.png` | 建築 UI（registry 已登記） |
| 其餘 `PX2D_icon_*_32` | 見 `IconSpriteRegistry.ts` |

```bash
bash client/scripts/sync-wip-icon-textures.sh
```

顯示：`IconSpriteRegistry.getSpriteFrame('PX2D_icon_res_food_32')`；**32→64→96…** 整數倍 `iconSize`。

## 匯入設定（全目錄）

- **filterMode**：`Point` / **Nearest**（各 Registry + `PixelSpriteUtil.applyPixelArtSampling`）
- **mipmap**：關閉
- **縮放**：禁止非整數倍拉伸像素畫

正式 `art/2d/**` 流程見 [`../../../../art/2d/README.md`](../../../../art/2d/README.md)；**勿**將 `_wip` 提升為正式夾。
