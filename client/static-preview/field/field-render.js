import {
  GRID_W,
  GRID_H,
  TILE_W,
  TILE_H,
  gridToScreen,
  computeOrigin,
} from './field-iso.js';

const TERRAIN_COLORS = [
  '#2a3d4a',
  '#324838',
  '#3a3a50',
  '#2d4048',
  '#353530',
  '#283842',
  '#3d2f28',
];

function terrainColor(gx, gy) {
  const i = (gx * 7 + gy * 11 + (gx ^ gy)) % TERRAIN_COLORS.length;
  return TERRAIN_COLORS[i];
}

function drawDiamond(ctx, cx, cy, fill, stroke) {
  ctx.beginPath();
  ctx.moveTo(cx, cy);
  ctx.lineTo(cx + TILE_W / 2, cy + TILE_H / 2);
  ctx.lineTo(cx, cy + TILE_H);
  ctx.lineTo(cx - TILE_W / 2, cy + TILE_H / 2);
  ctx.closePath();
  ctx.fillStyle = fill;
  ctx.fill();
  if (stroke) {
    ctx.strokeStyle = stroke;
    ctx.lineWidth = 1;
    ctx.stroke();
  }
}

const TYPE_COLORS = {
  infantry: { player: '#3d8b5a', enemy: '#a04040' },
  archer: { player: '#5070a0', enemy: '#904848' },
  cavalry: { player: '#a07050', enemy: '#883838' },
};

const TYPE_LABEL = {
  infantry: '步',
  archer: '弓',
  cavalry: '騎',
};

export function drawBattlefield(ctx, canvas, state) {
  const { originX, originY } = computeOrigin(canvas.width, canvas.height);
  ctx.clearRect(0, 0, canvas.width, canvas.height);

  for (let gy = 0; gy < GRID_H; gy++) {
    for (let gx = 0; gx < GRID_W; gx++) {
      const { x, y } = gridToScreen(gx, gy, originX, originY);
      const top = { x, y };
      let stroke = '#1a2530';
      let fill = terrainColor(gx, gy);

      const hl = state.highlightCells?.find((c) => c.x === gx && c.y === gy);
      if (hl) {
        fill = hl.kind === 'move' ? '#2a5a3a' : fill;
        stroke = '#5a9a6a';
      }

      drawDiamond(ctx, top.x, top.y, fill, stroke);
    }
  }

  const drawOrder = [...state.units].sort((a, b) => a.y + a.x - (b.y + b.x));
  for (const u of drawOrder) {
    drawUnitStack(ctx, u, originX, originY, state.selectedId === u.id);
  }

  return { originX, originY };
}

function drawUnitStack(ctx, unit, originX, originY, selected) {
  const { x, y } = gridToScreen(unit.x, unit.y, originX, originY);
  const cx = x;
  const cy = y + TILE_H / 2 + 4;
  const side = unit.side;
  const palette = TYPE_COLORS[unit.type]?.[side] ?? '#888';
  const w = 28;
  const h = 22;

  if (selected) {
    ctx.strokeStyle = '#f0d040';
    ctx.lineWidth = 3;
    ctx.strokeRect(cx - w / 2 - 3, cy - h / 2 - 3, w + 6, h + 6);
  }

  ctx.fillStyle = palette;
  ctx.fillRect(cx - w / 2, cy - h / 2, w, h);
  ctx.strokeStyle = side === 'player' ? '#1a4080' : '#601818';
  ctx.lineWidth = 2;
  ctx.strokeRect(cx - w / 2, cy - h / 2, w, h);

  ctx.fillStyle = '#fff';
  ctx.font = 'bold 11px system-ui, sans-serif';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText(TYPE_LABEL[unit.type] ?? '?', cx, cy);

  ctx.fillStyle = '#ffe066';
  ctx.font = 'bold 12px system-ui, sans-serif';
  ctx.fillText(String(unit.troops), cx, cy - h / 2 - 10);

  if (unit.troops <= 0) {
    ctx.fillStyle = 'rgba(0,0,0,0.5)';
    ctx.fillRect(cx - w / 2, cy - h / 2, w, h);
    ctx.fillStyle = '#aaa';
    ctx.font = '10px system-ui';
    ctx.fillText('潰', cx, cy);
  }
}
