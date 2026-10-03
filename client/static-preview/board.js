import { BOARD_SIZE, LOCAL_PLAYER_OWNER } from './config.js';
import { boardIsoTileDisplaySize, loadTerrainTileImages } from './tile-images.js';
import { boardUnitDisplaySize, loadUnitImages } from './unit-images.js';

const OWNER_COLORS = {
  0: { fill: '#3d7ee8', stroke: '#1a4a9e', label: 'P0' },
  1: { fill: '#e84d3d', stroke: '#9e1a1a', label: 'P1' },
};

const TYPE_LABEL = {
  0: '步',
  1: '弓',
  2: '騎',
};

const TERRAIN_TINT = {
  0: '#2a3344',
  1: '#4a4038',
  2: '#2d4a32',
  3: '#1e3a52',
  4: '#3d3a50',
  5: '#454040',
};

export class TacticalBoardRenderer {
  constructor(canvas) {
    this.canvas = canvas;
    this.ctx = canvas.getContext('2d');
    this.cellPx = 28;
    this.padding = 12;
    this.selectedUnitId = null;
    this.legalCells = [];
    /** @type {Map<number, HTMLImageElement>} */
    this.terrainTiles = new Map();
    /** @type {Map<number, HTMLImageElement>} */
    this.unitSprites = new Map();
    this._lastSnap = null;
    this.resize();
    void loadTerrainTileImages().then((tiles) => {
      this.terrainTiles = tiles;
      if (this._lastSnap) {
        this.draw(this._lastSnap);
      }
    });
    void loadUnitImages().then((units) => {
      this.unitSprites = units;
      if (this._lastSnap) {
        this.draw(this._lastSnap);
      }
    });
  }

  resize() {
    const size = this.padding * 2 + BOARD_SIZE * this.cellPx;
    this.canvas.width = size;
    this.canvas.height = size;
  }

  cellAtPixel(px, py) {
    const x = Math.floor((px - this.padding) / this.cellPx);
    const y = Math.floor((py - this.padding) / this.cellPx);
    if (x < 0 || y < 0 || x >= BOARD_SIZE || y >= BOARD_SIZE) {
      return null;
    }
    return { x, y };
  }

  setSelection(unitId, legalCells) {
    this.selectedUnitId = unitId;
    this.legalCells = legalCells ?? [];
  }

  clearSelection() {
    this.selectedUnitId = null;
    this.legalCells = [];
  }

  drawTerrainTile(ctx, terrain, px, py, cellPx) {
    const img =
      this.terrainTiles.get(terrain) ??
      this.terrainTiles.get(0);
    if (!img?.complete || !img.naturalWidth) {
      return false;
    }
    const { width: dw, height: dh } = boardIsoTileDisplaySize(cellPx);
    const dx = px + (cellPx - dw) / 2;
    const dy = py + (cellPx - dh) / 2;
    ctx.imageSmoothingEnabled = false;
    ctx.drawImage(img, 0, 0, img.naturalWidth, img.naturalHeight, dx, dy, dw, dh);
    return true;
  }

  draw(snap) {
    this._lastSnap = snap;
    const ctx = this.ctx;
    const { cellPx, padding } = this;
    ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
    ctx.fillStyle = '#12161f';
    ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);

    const legalSet = new Set(this.legalCells.map((c) => `${c.x},${c.y}`));

    for (let y = 0; y < BOARD_SIZE; y++) {
      for (let x = 0; x < BOARD_SIZE; x++) {
        const cell = snap.cells[y][x];
        const px = padding + x * cellPx;
        const py = padding + y * cellPx;
        const drewTile = this.drawTerrainTile(ctx, cell.terrain, px, py, cellPx);
        if (!drewTile) {
          ctx.fillStyle = TERRAIN_TINT[cell.terrain] ?? TERRAIN_TINT[0];
          ctx.fillRect(px + 1, py + 1, cellPx - 2, cellPx - 2);
        }
        if (!cell.passable) {
          ctx.fillStyle = 'rgba(0,0,0,0.35)';
          ctx.fillRect(px + 1, py + 1, cellPx - 2, cellPx - 2);
        }
        if (legalSet.has(`${x},${y}`)) {
          ctx.fillStyle = 'rgba(80, 200, 120, 0.45)';
          ctx.fillRect(px + 2, py + 2, cellPx - 4, cellPx - 4);
        }
      }
    }

    for (const unit of snap.units) {
      if (unit.hp <= 0) {
        continue;
      }
      const cx = padding + unit.x * cellPx + cellPx / 2;
      const cy = padding + unit.y * cellPx + cellPx / 2;
      const pal = OWNER_COLORS[unit.owner] ?? { fill: '#888', stroke: '#444', label: '?' };
      const r = cellPx * 0.36;
      const unitImg = this.unitSprites.get(unit.type) ?? this.unitSprites.get(0);
      const drewSprite =
        unitImg?.complete &&
        unitImg.naturalWidth > 0 &&
        this.drawUnitSprite(ctx, unitImg, cx, cy, cellPx, unit.owner, pal);
      if (!drewSprite) {
        ctx.beginPath();
        ctx.arc(cx, cy, r, 0, Math.PI * 2);
        ctx.fillStyle = pal.fill;
        ctx.fill();
        ctx.lineWidth = unit.id === this.selectedUnitId ? 3 : 2;
        ctx.strokeStyle = unit.id === this.selectedUnitId ? '#ffe566' : pal.stroke;
        ctx.stroke();
        ctx.fillStyle = '#f5f5f5';
        ctx.font = `bold ${Math.floor(cellPx * 0.38)}px system-ui,sans-serif`;
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(TYPE_LABEL[unit.type] ?? '?', cx, cy);
      } else if (unit.id === this.selectedUnitId) {
        ctx.lineWidth = 3;
        ctx.strokeStyle = '#ffe566';
        ctx.beginPath();
        ctx.arc(cx, cy, r + 2, 0, Math.PI * 2);
        ctx.stroke();
      }
      if (unit.owner === LOCAL_PLAYER_OWNER) {
        ctx.strokeStyle = 'rgba(255,255,255,0.25)';
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.arc(cx, cy, r + 3, 0, Math.PI * 2);
        ctx.stroke();
      }
    }
  }

  drawUnitSprite(ctx, img, cx, cy, cellPx, owner, pal) {
    const { width: dw, height: dh } = boardUnitDisplaySize(cellPx);
    const dx = cx - dw / 2;
    const dy = cy - dh / 2;
    ctx.save();
    ctx.imageSmoothingEnabled = false;
    if (owner !== LOCAL_PLAYER_OWNER) {
      ctx.filter = 'hue-rotate(-18deg) saturate(1.15)';
    }
    ctx.drawImage(img, 0, 0, img.naturalWidth, img.naturalHeight, dx, dy, dw, dh);
    ctx.restore();
    ctx.save();
    ctx.globalAlpha = 0.35;
    ctx.fillStyle = pal.fill;
    ctx.fillRect(dx, dy + dh - 4, dw, 4);
    ctx.restore();
    return true;
  }
}
