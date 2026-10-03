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

function drawUnitStack(ctx, unit, originX, originY) {
  const layout = stackLayout(unit.x, unit.y, originX, originY);
  const { anchorX, left, bodyTop, bodyDrawH, drawW } = layout;
  const stackImg = getStackImage(unit.type);
  const bodySrcH = stackImg?.naturalHeight
    ? stackImg.naturalHeight - STACK_SRC_CROP_TOP
    : 224 - STACK_SRC_CROP_TOP;

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
