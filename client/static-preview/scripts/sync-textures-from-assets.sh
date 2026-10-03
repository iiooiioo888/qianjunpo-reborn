#!/usr/bin/env bash
# Copy into static-preview for /qjp/ deploy (read-only sources; never modify art/).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
ASSETS="${ROOT}/client/assets/resources/textures/2d"
DST="${ROOT}/client/static-preview/textures/2d"
CHARS_PREVIEW="${ROOT}/client/static-preview/chars"

mkdir -p "${DST}/tiles" "${DST}/chars" "${DST}/units" "${DST}/icons" "${CHARS_PREVIEW}"

if [[ -d "${ASSETS}/tiles" ]]; then
  cp -f "${ASSETS}/tiles/"*.png "${DST}/tiles/" 2>/dev/null || true
fi
if [[ -d "${ASSETS}/units" ]]; then
  cp -f "${ASSETS}/units/"*.png "${DST}/units/" 2>/dev/null || true
fi
if [[ -d "${ASSETS}/icons" ]]; then
  cp -f "${ASSETS}/icons/"*.png "${DST}/icons/" 2>/dev/null || true
fi

for f in \
  PX2D_CHAR_WEI_Caocao_ex_v04.png \
  PX2D_CHAR_SHU_Zhangfei_ex_v04.png \
  PX2D_CHAR_WU_Placeholder_01_v04.png
do
  if [[ -f "${ASSETS}/chars/${f}" ]]; then
    cp -f "${ASSETS}/chars/${f}" "${CHARS_PREVIEW}/${f}"
    cp -f "${ASSETS}/chars/${f}" "${DST}/chars/${f}"
  fi
done

echo "Synced -> ${DST} and ${CHARS_PREVIEW}"
