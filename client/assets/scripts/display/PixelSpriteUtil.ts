import { SpriteFrame, Texture2D } from 'cc';

/** Cartoon pixel风：Nearest / Point，關閉 mipmap 語意（Cocos 2D 預設無 mipmap）。 */
export function applyPixelArtSampling(frame: SpriteFrame | null): void {
  if (!frame?.texture) {
    return;
  }
  const tex = frame.texture;
  tex.setFilters(Texture2D.Filter.NEAREST, Texture2D.Filter.NEAREST);
}

/**
 * v03 過審後單位貼圖畫布（像素）。v02 仍由 registry 指向既有檔名；顯示時用 {@link boardUnitDisplaySize} 整數倍縮放。
 */
export const UNIT_ART_CANVAS_PX = 128;

/** v03 角色卡畫布（像素）；寬 × 高。 */
export const CHAR_CARD_ART_WIDTH_PX = 320;
export const CHAR_CARD_ART_HEIGHT_PX = 400;

/**
 * 棋盤格內單位 Sprite 邊長：以畫布像素為基準，僅允許整數倍放大（1…maxUpscale）或反覆 ÷2 縮小，避免非整數縮放。
 */
export function boardUnitDisplaySize(
  artCanvasPx: number,
  cellSizePx: number,
  maxUpscale = 2,
): number {
  for (let k = maxUpscale; k >= 1; k--) {
    const scaled = artCanvasPx * k;
    if (scaled <= cellSizePx) {
      return scaled;
    }
  }
  let size = artCanvasPx;
  while (size > cellSizePx) {
    if (size % 2 !== 0) {
      break;
    }
    size /= 2;
  }
  return Math.min(size, cellSizePx);
}

/**
 * 角色卡 UI：在 maxW×maxH 框內，對 320×400 畫布取共同整數倍（僅縮小時 ÷2 鏈或 1×），保持寬高比。
 */
export function charCardDisplaySize(
  maxWidth: number,
  maxHeight: number,
  artW = CHAR_CARD_ART_WIDTH_PX,
  artH = CHAR_CARD_ART_HEIGHT_PX,
  maxUpscale = 2,
): { width: number; height: number } {
  let scale = maxUpscale;
  while (scale >= 1) {
    const w = artW * scale;
    const h = artH * scale;
    if (w <= maxWidth && h <= maxHeight) {
      return { width: w, height: h };
    }
    scale -= 1;
  }
  let w = artW;
  let h = artH;
  while (w > maxWidth || h > maxHeight) {
    if (w % 2 !== 0 || h % 2 !== 0) {
      break;
    }
    w /= 2;
    h /= 2;
  }
  return { width: Math.min(w, maxWidth), height: Math.min(h, maxHeight) };
}
