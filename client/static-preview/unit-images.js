import { UNIT_ART_CANVAS_PX, UNIT_TEXTURE_SRC } from './asset-registry.js';
import { resolveAppUrl } from './paths.js';

/**
 * Integer-scale unit sprite into a board cell (Nearest), mirrors PixelSpriteUtil.boardUnitDisplaySize.
 */
export function boardUnitDisplaySize(cellSizePx, maxUpscale = 2) {
  for (let k = maxUpscale; k >= 1; k--) {
    const w = UNIT_ART_CANVAS_PX * k;
    const h = UNIT_ART_CANVAS_PX * k;
    if (w <= cellSizePx && h <= cellSizePx) {
      return { width: w, height: h };
    }
  }
  let w = UNIT_ART_CANVAS_PX;
  let h = UNIT_ART_CANVAS_PX;
  while (w > cellSizePx || h > cellSizePx) {
    if (w % 2 !== 0 || h % 2 !== 0) {
      break;
    }
    w /= 2;
    h /= 2;
  }
  const fittedW = Math.min(w, cellSizePx);
  const fittedH = Math.min(h, cellSizePx);
  if (fittedW < cellSizePx || fittedH < cellSizePx) {
    return { width: cellSizePx, height: cellSizePx };
  }
  return { width: fittedW, height: fittedH };
}

/**
 * @returns {Promise<Map<number, HTMLImageElement>>} keyed by unit type
 */
export async function loadUnitImages() {
  const map = new Map();
  const byPath = new Map();
  for (const [typeStr, relPath] of Object.entries(UNIT_TEXTURE_SRC)) {
    const type = Number(typeStr);
    if (byPath.has(relPath)) {
      const existing = byPath.get(relPath);
      map.set(type, existing);
      continue;
    }
    await new Promise((resolve) => {
      const img = new Image();
      img.decoding = 'async';
      img.onload = () => {
        if (img.naturalWidth > 0) {
          map.set(type, img);
          byPath.set(relPath, img);
        }
        resolve();
      };
      img.onerror = () => resolve();
      img.src = resolveAppUrl(relPath);
    });
    if (byPath.has(relPath)) {
      map.set(type, byPath.get(relPath));
    }
  }
  return map;
}
