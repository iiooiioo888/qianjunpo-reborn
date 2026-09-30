import { Color, Graphics } from 'cc';
import { Coord } from '../logic/BoardCoord';
import { ViewSnapshot } from '../logic/TacticalSnapshot';

/** Shared mock board selection colors (display-only). */
export const SELECTION_RING_STROKE = new Color(255, 220, 60, 255);
export const SELECTION_RING_GLOW = new Color(255, 200, 40, 90);
export const SELECTED_CELL_FILL = new Color(255, 220, 60, 45);
export const LEGAL_CELL_FILL = new Color(72, 190, 110, 95);
export const LEGAL_CELL_STROKE = new Color(120, 235, 150, 200);

export function cellPixelOrigin(snap: ViewSnapshot, c: Coord, cellSize: number): { px: number; py: number } {
  return {
    px: c.x * cellSize,
    py: (snap.boardSize - 1 - c.y) * cellSize,
  };
}

/** Corner bracket markers for legal move destinations. */
export function drawLegalMoveCell(g: Graphics, px: number, py: number, cellSize: number): void {
  const inset = 4;
  const inner = cellSize - inset * 2;
  g.fillColor = LEGAL_CELL_FILL;
  g.rect(px + inset, py + inset, inner, inner);
  g.fill();
  g.strokeColor = LEGAL_CELL_STROKE;
  g.lineWidth = 1.5;
  g.rect(px + inset + 0.5, py + inset + 0.5, inner - 1, inner - 1);
  g.stroke();

  const tick = Math.min(7, Math.floor(cellSize * 0.22));
  const x0 = px + inset;
  const y0 = py + inset;
  const x1 = px + cellSize - inset;
  const y1 = py + cellSize - inset;
  g.lineWidth = 2;
  g.strokeColor = LEGAL_CELL_STROKE;
  const corners: [number, number, number, number, number, number][] = [
    [x0, y0 + tick, x0, y0, x0 + tick, y0],
    [x1 - tick, y0, x1, y0, x1, y0 + tick],
    [x0, y1 - tick, x0, y1, x0 + tick, y1],
    [x1 - tick, y1, x1, y1, x1, y1 - tick],
  ];
  for (const seg of corners) {
    g.moveTo(seg[0], seg[1]);
    g.lineTo(seg[2], seg[3]);
    g.lineTo(seg[4], seg[5]);
    g.stroke();
  }
}

export function drawSelectedUnitCell(g: Graphics, px: number, py: number, cellSize: number): void {
  g.fillColor = SELECTED_CELL_FILL;
  g.rect(px + 1, py + 1, cellSize - 2, cellSize - 2);
  g.fill();
}

/** HUD 角色卡選取高亮（與單位選取環同色系）。 */
export function drawSelectedCharCardFrame(g: Graphics, width: number, height: number): void {
  g.lineWidth = 2.5;
  g.strokeColor = SELECTION_RING_STROKE;
  g.rect(1, 1, width - 2, height - 2);
  g.stroke();
  g.lineWidth = 1;
  g.strokeColor = SELECTION_RING_GLOW;
  g.rect(0.5, 0.5, width - 1, height - 1);
  g.stroke();
}

export function drawUnitSelectionRing(g: Graphics, cellSize: number): void {
  const cx = cellSize / 2;
  const cy = cellSize / 2;
  const rOuter = cellSize / 2 - 2;
  const rInner = cellSize / 2 - 5;
  g.lineWidth = 3;
  g.strokeColor = SELECTION_RING_STROKE;
  g.circle(cx, cy, rInner);
  g.stroke();
  g.lineWidth = 1;
  g.strokeColor = SELECTION_RING_GLOW;
  g.circle(cx, cy, rOuter);
  g.stroke();
}
