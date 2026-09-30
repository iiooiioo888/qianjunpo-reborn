import { BOARD_SIZE } from './TacticalSnapshot';

export interface Coord {
  x: number;
  y: number;
}

export function inBounds(c: Coord): boolean {
  return c.x >= 0 && c.x < BOARD_SIZE && c.y >= 0 && c.y < BOARD_SIZE;
}

/** Chebyshev distance (8-neighbor steps), aligned with pkg/board.Chebyshev. */
export function chebyshev(a: Coord, b: Coord): number {
  const dx = Math.abs(a.x - b.x);
  const dy = Math.abs(a.y - b.y);
  return Math.max(dx, dy);
}

const NEIGHBOR_DELTAS: ReadonlyArray<Coord> = [
  { x: 0, y: -1 },
  { x: 1, y: -1 },
  { x: 1, y: 0 },
  { x: 1, y: 1 },
  { x: 0, y: 1 },
  { x: -1, y: 1 },
  { x: -1, y: 0 },
  { x: -1, y: -1 },
];

export function neighbors(c: Coord): Coord[] {
  const out: Coord[] = [];
  for (const d of NEIGHBOR_DELTAS) {
    const nc = { x: c.x + d.x, y: c.y + d.y };
    if (inBounds(nc)) {
      out.push(nc);
    }
  }
  return out;
}

export function coordKey(c: Coord): string {
  return `${c.x},${c.y}`;
}
