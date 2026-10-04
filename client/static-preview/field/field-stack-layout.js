import { TILE_H, gridToScreen } from './field-iso.js';

/** 192×224 堆圖在畫布上的寬度（舊版整堆繪製參考，點選改用 figures 外框） */
export const STACK_DRAW_W = 52;

/** 原稿頂部佔位兵力帶（步／弓／騎綠字至約 y 54–55），繪製與點選皆從此裁切 */
export const STACK_SRC_CROP_TOP = 56;
export const STACK_NATURAL_W = 192;
export const STACK_NATURAL_H = 224;

/** 單兵在畫布上的目標寬度 */
export const FIGURE_DRAW_W = 22;

/** 相鄰個體外框（寬度 FIGURE_DRAW_W）之間的水平留白 */
export const FIGURE_FRAME_GAP_PX = 6;

/** 相鄰腳點水平距 = 外框寬 + 留白 */
export const FIGURE_FOOT_SPACING = FIGURE_DRAW_W + FIGURE_FRAME_GAP_PX;

/**
 * 從兵堆稿裁切前排三兵（座標相對於裁切後兵身，y=0 為 STACK_SRC_CROP_TOP 之下）
 * @type {Record<'infantry'|'archer'|'cavalry', Array<{ sx: number, sy: number, sw: number, sh: number }>>}
 */
export const FIGURE_SRC_CROPS = {
  infantry: [
    { sx: 36, sy: 88, sw: 38, sh: 72 },
    { sx: 78, sy: 82, sw: 38, sh: 78 },
    { sx: 120, sy: 88, sw: 36, sh: 72 },
  ],
  archer: [
    { sx: 10, sy: 70, sw: 40, sh: 85 },
    { sx: 55, sy: 65, sw: 42, sh: 90 },
    { sx: 105, sy: 70, sw: 45, sh: 85 },
  ],
  cavalry: [
    { sx: 8, sy: 55, sw: 48, sh: 95 },
    { sx: 58, sy: 50, sw: 50, sh: 100 },
    { sx: 115, sy: 60, sw: 48, sh: 90 },
  ],
};

/** 格內三兵腳點偏移（相對 anchor；左右各距中心 FIGURE_FOOT_SPACING） */
export const FIGURE_FOOT_OFFSETS = [
  { dx: -FIGURE_FOOT_SPACING, dy: 0 },
  { dx: 0, dy: -4 },
  { dx: FIGURE_FOOT_SPACING, dy: 0 },
];

export const SIDE_TINT = {
  player: 'rgba(70, 130, 255, 0.28)',
  enemy: 'rgba(255, 70, 70, 0.32)',
};

export const SIDE_FLAG_FILL = {
  player: '#2d6cdf',
  enemy: '#c0392b',
};

export function stackLayout(gx, gy, originX, originY) {
  const { x, y } = gridToScreen(gx, gy, originX, originY);
  const anchorX = x;
  const anchorY = y + TILE_H / 2 + 2;

  const figures = FIGURE_FOOT_OFFSETS.map(({ dx, dy }) => ({
    cx: anchorX + dx,
    footY: anchorY + dy,
  }));

  let pickLeft = Infinity;
  let pickTop = Infinity;
  let pickRight = -Infinity;
  let pickBottom = -Infinity;

  for (const f of figures) {
    const left = f.cx - FIGURE_DRAW_W / 2;
    const top = f.footY - Math.round(FIGURE_DRAW_W * 1.35);
    pickLeft = Math.min(pickLeft, left);
    pickTop = Math.min(pickTop, top);
    pickRight = Math.max(pickRight, left + FIGURE_DRAW_W);
    pickBottom = Math.max(pickBottom, f.footY + 2);
  }

  const flagX = anchorX - 24;
  const flagTop = anchorY - 22;
  pickLeft = Math.min(pickLeft, flagX - 2);
  pickTop = Math.min(pickTop, flagTop - 2);
  pickBottom = Math.max(pickBottom, anchorY + 2);

  const bodyTop = pickTop;
  const bodyDrawH = pickBottom - pickTop;
  const left = pickLeft;
  const drawW = pickRight - pickLeft;

  return {
    anchorX,
    anchorY,
    figures,
    flagX,
    flagTop,
    left,
    drawW,
    bodyTop,
    bodyDrawH,
    pickLeft,
    pickTop,
    pickRight,
    pickBottom,
  };
}

/** 與格內個體＋旗幟同一外框 */
export function pointInStackPick(px, py, layout) {
  return (
    px >= layout.pickLeft &&
    px <= layout.pickRight &&
    py >= layout.pickTop &&
    py <= layout.pickBottom
  );
}
