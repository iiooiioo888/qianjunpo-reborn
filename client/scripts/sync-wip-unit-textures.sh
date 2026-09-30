#!/usr/bin/env bash
# 自 art/2d/_wip/units 複製 STANDARD 單位貼圖至 Cocos resources（不修改 art/ 源目錄、不提升 _wip）。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SRC="${ROOT}/art/2d/_wip/units"
DST="${ROOT}/client/assets/resources/textures/2d/units"
mkdir -p "${DST}"
for f in PX2D_unit_infantry.png PX2D_unit_cavalry.png; do
  if [[ ! -f "${SRC}/${f}" ]]; then
    echo "missing ${SRC}/${f} — git checkout art/2d/_wip/units 或從 main 拉資產（art #48 STANDARD）" >&2
    exit 1
  fi
  cp -f "${SRC}/${f}" "${DST}/${f}"
  echo "copied ${f} -> ${DST}/"
done
echo "Done. 在 Cocos 中對 textures/2d/units 右鍵重新導入，確認 filterMode=Nearest（UnitSpriteRegistry + PixelSpriteUtil 執行期亦會 setFilters NEAREST）。"
