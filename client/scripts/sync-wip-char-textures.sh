#!/usr/bin/env bash
# 自 art/2d/_wip/characters 複製 v02 角色卡至 Cocos resources（不修改 art/ 源目錄、不提升 _wip）。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SRC="${ROOT}/art/2d/_wip/characters"
DST="${ROOT}/client/assets/resources/textures/2d/chars"
mkdir -p "${DST}"
for f in \
  PX2D_CHAR_WEI_Caocao_ex_v02.png \
  PX2D_CHAR_SHU_Zhangfei_ex_v02.png \
  PX2D_CHAR_WU_Placeholder_01_v02.png
do
  if [[ ! -f "${SRC}/${f}" ]]; then
    echo "missing ${SRC}/${f} — git checkout art/2d/_wip/characters 或從 main 拉資產" >&2
    exit 1
  fi
  cp -f "${SRC}/${f}" "${DST}/${f}"
  echo "copied ${f} -> ${DST}/"
done
echo "Done. 在 Cocos 中對 textures/2d/chars 右鍵重新導入；registry 鍵見 CharacterCardSpriteRegistry.ts。"
