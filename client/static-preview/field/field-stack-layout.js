import { TILE_H, gridToScreen } from './field-iso.js';

/** 192×224 堆圖在畫布上的寬度 */
export const STACK_DRAW_W = 52;

/** 原稿頂部佔位兵力帶（約 y 15–53），繪製時裁掉、點選仍含此高度 */
export const STACK_SRC_CROP_TOP = 53;
export const STACK_NATURAL_W = 192;
export const STACK_NATURAL_H = 224;

export function stackLayout(gx, gy, originX, originY) {
  const { x, y } = gridToScreen(gx, gy, originX, originY);
  const anchorX = x;
  const anchorY = y + TILE_H / 2 + 2;
  const drawW = STACK_DRAW_W;
  const fullDrawH = Math.round(STACK_DRAW_W * (STACK_NATURAL_H / STACK_NATURAL_W));
  const bodySrcH = STACK_NATURAL_H - STACK_SRC_CROP_TOP;
  const bodyDrawH = Math.round(STACK_DRAW_W * (bodySrcH / STACK_NATURAL_W));
  const left = anchorX - drawW / 2;
  const bodyTop = anchorY - bodyDrawH;
  const pickTop = anchorY - fullDrawH;
  return {
    anchorX,
    anchorY,
    left,
    drawW,
    bodyTop,
    bodyDrawH,
    pickTop,
    pickH: fullDrawH,
  };
}

export function pointInStackPick(px, py, layout) {
  return (
    px >= layout.left &&
    px <= layout.left + layout.drawW &&
    py >= layout.pickTop &&
    py <= layout.anchorY + 4
  );
}
