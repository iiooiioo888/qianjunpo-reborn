import { TILE_H, gridToScreen } from './field-iso.js';

/** 192×224 堆圖在畫布上的寬度 */
export const STACK_DRAW_W = 52;

/** 原稿頂部佔位兵力帶（步／弓／騎綠字至約 y 54–55），繪製與點選皆從此裁切 */
export const STACK_SRC_CROP_TOP = 56;
export const STACK_NATURAL_W = 192;
export const STACK_NATURAL_H = 224;

export function stackLayout(gx, gy, originX, originY) {
  const { x, y } = gridToScreen(gx, gy, originX, originY);
  const anchorX = x;
  const anchorY = y + TILE_H / 2 + 2;
  const drawW = STACK_DRAW_W;
  const bodySrcH = STACK_NATURAL_H - STACK_SRC_CROP_TOP;
  const bodyDrawH = Math.round(STACK_DRAW_W * (bodySrcH / STACK_NATURAL_W));
  const left = anchorX - drawW / 2;
  const bodyTop = anchorY - bodyDrawH;
  return {
    anchorX,
    anchorY,
    left,
    drawW,
    bodyTop,
    bodyDrawH,
  };
}

/** 與 drawImage 裁切後的兵身同一矩形 */
export function pointInStackPick(px, py, layout) {
  return (
    px >= layout.left &&
    px <= layout.left + layout.drawW &&
    py >= layout.bodyTop &&
    py <= layout.anchorY + 4
  );
}
