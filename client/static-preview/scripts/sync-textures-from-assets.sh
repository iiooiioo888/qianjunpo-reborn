#!/usr/bin/env bash
# Copy 2D textures into static-preview/textures/2d for deploy-web-preview (/qjp/).
# Sources (read-only): client/assets/resources/textures/2d and art/2d/_wip/characters (v04 cards).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
ASSETS="${ROOT}/client/assets/resources/textures/2d"
ART_CHARS="${ROOT}/art/2d/_wip/characters"
DST="${ROOT}/client/static-preview/textures/2d"

mkdir -p "${DST}/tiles" "${DST}/chars" "${DST}/units" "${DST}/icons"

if [[ -d "${ASSETS}/tiles" ]]; then
  cp -f "${ASSETS}/tiles/"*.png "${DST}/tiles/" 2>/dev/null || true
fi
if [[ -d "${ASSETS}/units" ]]; then
  cp -f "${ASSETS}/units/"*.png "${DST}/units/" 2>/dev/null || true
fi
if [[ -d "${ASSETS}/icons" ]]; then
  cp -f "${ASSETS}/icons/"*.png "${DST}/icons/" 2>/dev/null || true
fi
if [[ -d "${ASSETS}/chars" ]]; then
  cp -f "${ASSETS}/chars/"*.png "${DST}/chars/" 2>/dev/null || true
fi

for f in \
  PX2D_CHAR_WEI_Caocao_ex_v04.png \
  PX2D_CHAR_SHU_Zhangfei_ex_v04.png \
  PX2D_CHAR_WU_Placeholder_01_v04.png
do
  if [[ -f "${ART_CHARS}/${f}" ]]; then
    cp -f "${ART_CHARS}/${f}" "${DST}/chars/${f}"
  fi
done

echo "Synced textures -> ${DST}"
