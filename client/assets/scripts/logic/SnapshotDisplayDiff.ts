import { ViewSnapshot } from './TacticalSnapshot';

/** Stable key for unit-layer redraw (frame + alive unit pose/hp). */
export function snapshotUnitsRenderKey(snap: ViewSnapshot): string {
  const parts = snap.units
    .filter((u) => u.hp > 0)
    .map((u) => `${u.id}:${u.x},${u.y},${u.hp},${u.type}`)
    .sort()
    .join('|');
  return `${snap.lockstepFrame}#${parts}`;
}

export function shouldRedrawUnits(prev: ViewSnapshot | null, next: ViewSnapshot): boolean {
  if (!prev) {
    return true;
  }
  return snapshotUnitsRenderKey(prev) !== snapshotUnitsRenderKey(next);
}

/** Terrain grid is static for mock/live initial battles; skip grid if unchanged. */
export function shouldRedrawGrid(prev: ViewSnapshot | null, next: ViewSnapshot): boolean {
  if (!prev) {
    return true;
  }
  if (prev.boardSize !== next.boardSize) {
    return true;
  }
  for (let y = 0; y < next.boardSize; y++) {
    for (let x = 0; x < next.boardSize; x++) {
      const a = prev.cells[y][x];
      const b = next.cells[y][x];
      if (a.terrain !== b.terrain || a.passable !== b.passable) {
        return true;
      }
    }
  }
  return false;
}
