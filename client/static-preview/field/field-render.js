import {
  GRID_W,
  GRID_H,
  TILE_W,
  TILE_H,
  diamondVertices,
  gridToScreen,
  computeOrigin,
} from './field-iso.js';
import { getGroundImage, getStackImage } from './field-assets.js';
import {
  STACK_SRC_CROP_TOP,
  FIGURE_DRAW_W,
  FIGURE_SRC_CROPS,
  SIDE_FLAG_FILL,
  SIDE_TINT,
  stackLayout,
} from './field-stack-layout.js';

/** 單兵離屏染色用（避免主畫布 source-atop 染到草地／旗幟） */
let figureScratchCanvas = null;

function getFigureScratch(drawW, drawH) {
  if (!figureScratchCanvas) {
    figureScratchCanvas = document.createElement('canvas');
  }
  if (figureScratchCanvas.width < drawW) {
    figureScratchCanvas.width = drawW;
  }
  if (figureScratchCanvas.height < drawH) {
    figureScratchCanvas.height = drawH;
  }
  const sctx = figureScratchCanvas.getContext('2d');
  sctx.clearRect(0, 0, drawW, drawH);
  return { canvas: figureScratchCanvas, ctx: sctx };
}

function drawDiamondPath(ctx, cx, cy) {
  const verts = diamondVertices(cx, cy);
  ctx.beginPath();
  ctx.moveTo(verts[0].x, verts[0].y);
  for (let i = 1; i < verts.length; i++) {
    ctx.lineTo(verts[i].x, verts[i].y);
  }
  ctx.closePath();
}

function drawGrassTile(ctx, topX, topY) {
  const ground = getGroundImage();
  const cx = topX;
  const cy = topY;
  if (ground?.complete && ground.naturalWidth) {
    ctx.drawImage(ground, cx - TILE_W / 2, cy, TILE_W, TILE_H);
    return;
  }
  drawDiamondPath(ctx, cx, cy);
  ctx.fillStyle = '#324838';
  ctx.fill();
  ctx.strokeStyle = '#1a2530';
  ctx.lineWidth = 1;
  ctx.stroke();
}

export function drawBattlefield(ctx, canvas, state) {
  const { originX, originY } = computeOrigin(canvas.width, canvas.height);
  ctx.clearRect(0, 0, canvas.width, canvas.height);

  for (let gy = 0; gy < GRID_H; gy++) {
    for (let gx = 0; gx < GRID_W; gx++) {
      const { x, y } = gridToScreen(gx, gy, originX, originY);
      drawGrassTile(ctx, x, y);
    }
  }

  const drawOrder = state.units
    .filter((u) => u.troops > 0)
    .sort((a, b) => a.y + a.x - (b.y + b.x));
  for (const u of drawOrder) {
    drawUnitStack(ctx, u, originX, originY);
  }

  drawDamagePopups(ctx, state.popups ?? [], originX, originY);

  if (state.gameResult) {
    drawGameResultOverlay(ctx, canvas, state.gameResult);
  }

  return { originX, originY };
}

function drawGameResultOverlay(ctx, canvas, gameResult) {
  const label = gameResult === 'win' ? '勝' : '負';
  const cx = canvas.width / 2;
  const cy = canvas.height / 2;
  ctx.save();
  ctx.fillStyle = 'rgba(0, 0, 0, 0.45)';
  ctx.fillRect(0, 0, canvas.width, canvas.height);
  ctx.font =
    'bold 80px system-ui, "PingFang TC", "Microsoft JhengHei", sans-serif';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.lineWidth = 8;
  ctx.strokeStyle = 'rgba(0, 0, 0, 0.85)';
  ctx.strokeText(label, cx, cy);
  ctx.fillStyle = gameResult === 'win' ? '#ffe066' : '#ff6b6b';
  ctx.fillText(label, cx, cy);
  ctx.restore();
}

function drawSideFlag(ctx, layout, side) {
  const { flagX, flagTop, anchorY } = layout;
  const fill = SIDE_FLAG_FILL[side] ?? SIDE_FLAG_FILL.enemy;
  ctx.save();
  ctx.strokeStyle = '#2a2a2a';
  ctx.lineWidth = 1.5;
  ctx.beginPath();
  ctx.moveTo(flagX, anchorY);
  ctx.lineTo(flagX, flagTop);
  ctx.stroke();
  ctx.fillStyle = fill;
  ctx.beginPath();
  ctx.moveTo(flagX, flagTop);
  ctx.lineTo(flagX + 11, flagTop + 4);
  ctx.lineTo(flagX, flagTop + 8);
  ctx.closePath();
  ctx.fill();
  ctx.strokeStyle = 'rgba(0,0,0,0.45)';
  ctx.lineWidth = 1;
  ctx.stroke();
  ctx.restore();
}

function drawFigureSprite(ctx, stackImg, crop, footX, footY, tint) {
  const srcY = STACK_SRC_CROP_TOP + crop.sy;
  const drawW = FIGURE_DRAW_W;
  const drawH = Math.max(18, Math.round(drawW * (crop.sh / crop.sw)));
  const left = footX - drawW / 2;
  const top = footY - drawH;

  if (stackImg?.complete && stackImg.naturalWidth) {
    const { canvas: scratch, ctx: sctx } = getFigureScratch(drawW, drawH);
    sctx.drawImage(
      stackImg,
      crop.sx,
      srcY,
      crop.sw,
      crop.sh,
      0,
      0,
      drawW,
      drawH,
    );
    if (tint) {
      sctx.save();
      sctx.globalCompositeOperation = 'source-atop';
      sctx.fillStyle = tint;
      sctx.fillRect(0, 0, drawW, drawH);
      sctx.restore();
    }
    ctx.drawImage(scratch, 0, 0, drawW, drawH, left, top, drawW, drawH);
    return;
  }

  ctx.fillStyle = '#555';
  ctx.fillRect(left, top, drawW, drawH);
}

function drawUnitStack(ctx, unit, originX, originY) {
  const layout = stackLayout(unit.x, unit.y, originX, originY);
  const stackImg = getStackImage(unit.type);
  const crops = FIGURE_SRC_CROPS[unit.type] ?? FIGURE_SRC_CROPS.infantry;
  const tint = SIDE_TINT[unit.side] ?? SIDE_TINT.enemy;

  drawSideFlag(ctx, layout, unit.side);

  const figureCount = Math.min(crops.length, layout.figures.length);
  for (let i = 0; i < figureCount; i++) {
    const { cx, footY } = layout.figures[i];
    drawFigureSprite(ctx, stackImg, crops[i], cx, footY, tint);
  }

  drawTroopCount(ctx, layout.anchorX, layout.bodyTop, unit.troops);
}

function drawTroopCount(ctx, cx, stackTop, troops) {
  const label = String(troops);
  ctx.font = 'bold 13px system-ui, "PingFang TC", "Microsoft JhengHei", sans-serif';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'bottom';
  ctx.lineWidth = 3;
  ctx.strokeStyle = 'rgba(0, 0, 0, 0.75)';
  ctx.strokeText(label, cx, stackTop - 4);
  ctx.fillStyle = '#ffe066';
  ctx.fillText(label, cx, stackTop - 4);
}

function drawDamagePopups(ctx, popups, originX, originY) {
  for (const p of popups) {
    const layout = stackLayout(p.gx, p.gy, originX, originY);
    const rise = (48 - p.ttl) * 0.6;
    const alpha = Math.min(1, p.ttl / 24);
    const x = layout.anchorX;
    const y = layout.bodyTop - 18 - rise;
    ctx.save();
    ctx.globalAlpha = alpha;
    ctx.font = 'bold 16px system-ui, "PingFang TC", "Microsoft JhengHei", sans-serif';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.lineWidth = 4;
    ctx.strokeStyle = 'rgba(0, 0, 0, 0.8)';
    ctx.strokeText(p.text, x, y);
    ctx.fillStyle = '#ff4444';
    ctx.fillText(p.text, x, y);
    ctx.restore();
  }
}
