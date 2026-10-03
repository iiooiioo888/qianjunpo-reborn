import {
  GRID_W,
  GRID_H,
  TILE_W,
  diamondVertices,
  gridToScreen,
  computeOrigin,
} from './field-iso.js';
import { getGroundImage, getStackImage } from './field-assets.js';
import { STACK_SRC_CROP_TOP, stackLayout } from './field-stack-layout.js';

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

function drawHighlightOverlay(ctx, topX, topY, kind) {
  drawDiamondPath(ctx, topX, topY);
  if (kind === 'move') {
    ctx.fillStyle = 'rgba(42, 90, 58, 0.45)';
    ctx.fill();
    ctx.strokeStyle = '#5a9a6a';
    ctx.lineWidth = 1;
    ctx.stroke();
  }
}

export function drawBattlefield(ctx, canvas, state) {
  const { originX, originY } = computeOrigin(canvas.width, canvas.height);
  ctx.clearRect(0, 0, canvas.width, canvas.height);

  for (let gy = 0; gy < GRID_H; gy++) {
    for (let gx = 0; gx < GRID_W; gx++) {
      const { x, y } = gridToScreen(gx, gy, originX, originY);
      drawGrassTile(ctx, x, y);
      const hl = state.highlightCells?.find((c) => c.x === gx && c.y === gy);
      if (hl) {
        drawHighlightOverlay(ctx, x, y, hl.kind);
      }
    }
  }

  const drawOrder = [...state.units].sort((a, b) => a.y + a.x - (b.y + b.x));
  for (const u of drawOrder) {
    drawUnitStack(ctx, u, originX, originY, state.selectedId === u.id);
  }

  return { originX, originY };
}

function drawUnitStack(ctx, unit, originX, originY, selected) {
  const layout = stackLayout(unit.x, unit.y, originX, originY);
  const { anchorX, left, bodyTop, bodyDrawH, drawW } = layout;
  const stackImg = getStackImage(unit.type);
  const bodySrcH = stackImg?.naturalHeight
    ? stackImg.naturalHeight - STACK_SRC_CROP_TOP
    : 224 - STACK_SRC_CROP_TOP;

  if (selected) {
    ctx.strokeStyle = '#f0d040';
    ctx.lineWidth = 3;
    ctx.strokeRect(left - 4, bodyTop - 4, drawW + 8, bodyDrawH + 8);
  }

  if (stackImg?.complete && stackImg.naturalWidth) {
    ctx.drawImage(
      stackImg,
      0,
      STACK_SRC_CROP_TOP,
      stackImg.naturalWidth,
      bodySrcH,
      left,
      bodyTop,
      drawW,
      bodyDrawH,
    );
  } else {
    ctx.fillStyle = '#555';
    ctx.fillRect(left, bodyTop, drawW, bodyDrawH);
  }

  drawTroopCount(ctx, anchorX, bodyTop, unit.troops);

  if (unit.troops <= 0) {
    ctx.fillStyle = 'rgba(0,0,0,0.55)';
    ctx.fillRect(left, bodyTop, drawW, bodyDrawH);
    ctx.fillStyle = '#ddd';
    ctx.font = 'bold 12px system-ui, sans-serif';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText('潰', anchorX, bodyTop + bodyDrawH / 2);
  }
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
