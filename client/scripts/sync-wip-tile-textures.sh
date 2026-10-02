#!/usr/bin/env bash
# 自 art/25d/_wip/tiles 複製 STANDARD ISO25 地格至 Cocos resources 與 static-preview（不修改 art/）。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SRC="${ROOT}/art/25d/_wip/tiles"
DST_COCOS="${ROOT}/client/assets/resources/textures/2d/tiles"
DST_PREVIEW="${ROOT}/client/static-preview/tiles"
mkdir -p "${DST_COCOS}" "${DST_PREVIEW}"
for f in ISO25_tile_grass_v02.png ISO25_tile_mountain_v01.png; do
  if [[ ! -f "${SRC}/${f}" ]]; then
    echo "missing ${SRC}/${f} — git pull main（art #83 STANDARD）" >&2
    exit 1
  fi
  cp -f "${SRC}/${f}" "${DST_COCOS}/${f}"
  cp -f "${SRC}/${f}" "${DST_PREVIEW}/${f}"
  echo "copied ${f} -> ${DST_COCOS}/ and ${DST_PREVIEW}/"
done
echo "Done. Cocos：對 textures/2d/tiles 重新導入，filterMode=Nearest；static-preview 已可直接載入 tiles/*.png。"
