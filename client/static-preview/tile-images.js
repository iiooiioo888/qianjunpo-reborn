import { ISO25_TILE_ART_H, ISO25_TILE_ART_W } from './config.js';
import { TERRAIN_TILE_SRC } from './asset-registry.js';
import { resolveAppUrl } from './paths.js';

/**
 * 整數倍縮放 ISO25 地格（Nearest），與 Cocos boardIsoTileDisplaySize 對齊。
 * @returns {{ width: number, height: number }}
 */
export function boardIsoTileDisplaySize(cellSizePx, maxUpscale = 2) {
  for (let k = maxUpscale; k >= 1; k--) {
    const w = ISO25_TILE_ART_W * k;
    const h = ISO25_TILE_ART_H * k;
    if (w <= cellSizePx && h <= cellSizePx) {
      return { width: w, height: h };
    }
  }
  let w = ISO25_TILE_ART_W;
  let h = ISO25_TILE_ART_H;
  while (w > cellSizePx || h > cellSizePx) {
    if (w % 2 !== 0 || h % 2 !== 0) {
      break;
    }
    w /= 2;
    h /= 2;
  }
  const fittedW = Math.min(w, cellSizePx);
  const fittedH = Math.min(h, cellSizePx);
  if (fittedW < cellSizePx || fittedH < Math.floor(cellSizePx / 2)) {
    return { width: cellSizePx, height: Math.floor(cellSizePx / 2) };
  }
  return { width: fittedW, height: fittedH };
}

/**
 * @returns {Promise<Map<number, HTMLImageElement>>}
 */
export async function loadTerrainTileImages() {
  const map = new Map();
  const entries = Object.entries(TERRAIN_TILE_SRC);
  await Promise.all(
    entries.map(
      ([terrainId, relPath]) =>
        new Promise((resolve) => {
          const img = new Image();
          img.decoding = 'async';
          img.onload = () => {
            if (img.naturalWidth > 0) {
              map.set(Number(terrainId), img);
            }
            resolve();
          };
          img.onerror = () => resolve();
          img.src = resolveAppUrl(relPath);
        }),
    ),
  );
  return map;
}
