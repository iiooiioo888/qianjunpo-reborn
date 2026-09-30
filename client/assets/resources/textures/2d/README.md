# Cocos `resources/textures/2d`

戰術客戶端 2D 像素貼圖匯入位（**非** `art/` 正式管線）。

## 單位（已自 `_wip` 複製 v02）

| 檔案 | 用途 |
|------|------|
| `units/PX2D_unit_infantry_v02.png` | `ViewUnit.type` 步兵（0 等預設） |
| `units/PX2D_unit_cavalry_v02.png` | 騎兵（type `2`） |

來源：`art/2d/_wip/units/`（工作稿，**勿**移入 `art/2d/units/` 正式夾）。更新時執行：

```bash
bash client/scripts/sync-wip-unit-textures.sh
```

## 卡通像素風匯入設定

- **filterMode**：`Point` / **Nearest**（`UnitSpriteRegistry` + `PixelSpriteUtil` 執行期 `setFilters(NEAREST)`）
- **mipmap**：關閉（2D UI／Sprite 預設）
- **縮放**：棋盤 `cellSize=32`，Sprite `CUSTOM` 縮至格內；放大棋盤時請用 **整數倍** `cellSize`（32→64→96…）

## 圖標（已自 `_wip` 複製 PX2D_icon_*_32）

| 檔案 | 用途（`IconAssetId`） |
|------|------------------------|
| `icons/PX2D_icon_res_food_32.png` 等五資源 | HUD `ResourceIconHudStrip`、資源列 |
| `icons/PX2D_icon_bld_*_32.png` | 建築類 UI（registry 已登記） |
| `icons/PX2D_icon_army_32.png` | 軍團／編制 |
| `icons/PX2D_icon_formation_32.png` | 陣形 |
| `icons/PX2D_icon_battle_report_32.png` | 戰報 |

來源：`art/2d/_wip/icons/`（工作稿，**勿**移入 `art/2d/icons/` 正式夾）。更新時執行：

```bash
bash client/scripts/sync-wip-icon-textures.sh
```

顯示層載入：`IconSpriteRegistry.getSpriteFrame('PX2D_icon_res_food_32')` 或常數 `ICON_ASSET_IDS.resFood`；`PixelSpriteUtil.applyPixelArtSampling` → **Nearest**；UI 請用 **32→64→96…** 整數倍 `iconSize`。

角色卡 v03 前暫不匯入 `characters/`。
