#!/usr/bin/env bash
# 自 art/2d/_wip/icons 複製 P0 像素圖標至 Cocos resources（不修改 art/ 源目錄、不提升 _wip）。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SRC="${ROOT}/art/2d/_wip/icons"
DST="${ROOT}/client/assets/resources/textures/2d/icons"
mkdir -p "${DST}"
ICONS=(
  PX2D_icon_army_32.png
  PX2D_icon_battle_report_32.png
  PX2D_icon_bld_academy_32.png
  PX2D_icon_bld_barracks_32.png
  PX2D_icon_bld_keep_32.png
  PX2D_icon_bld_lumber_32.png
  PX2D_icon_bld_warehouse_32.png
  PX2D_icon_formation_32.png
  PX2D_icon_res_food_32.png
  PX2D_icon_res_gold_32.png
  PX2D_icon_res_iron_32.png
  PX2D_icon_res_stone_32.png
  PX2D_icon_res_wood_32.png
)
for f in "${ICONS[@]}"; do
  if [[ ! -f "${SRC}/${f}" ]]; then
    echo "missing ${SRC}/${f} — git checkout art/2d/_wip/icons 或從 main 拉資產" >&2
    exit 1
  fi
  cp -f "${SRC}/${f}" "${DST}/${f}"
  echo "copied ${f} -> ${DST}/"
done
echo "Done. 在 Cocos 中對 textures/2d/icons 右鍵重新導入，確認 filterMode=Nearest（IconSpriteRegistry + PixelSpriteUtil 執行期亦會 setFilters NEAREST）。"
