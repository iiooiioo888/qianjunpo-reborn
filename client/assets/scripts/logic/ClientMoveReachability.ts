import { Coord, coordKey, neighbors } from './BoardCoord';
import { ViewSnapshot } from './TacticalSnapshot';

/**
 * Default move budget for duel units (pkg/tactical defaultUnit Move: 4).
 * ViewSnapshot does not expose per-unit Move yet — display-only reachability.
 */
export const CLIENT_DEFAULT_MOVE_POINTS = 4;

function isWalkable(snap: ViewSnapshot, c: Coord, ignoreUnitId: number): boolean {
  const cell = snap.cells[c.y]?.[c.x];
  if (!cell || !cell.passable) {
    return false;
  }
  if (cell.unitId !== 0 && cell.unitId !== ignoreUnitId) {
    return false;
  }
  return true;
}

/**
 * BFS reachable destinations within movePoints (shortest path length).
 * Mirrors pkg/pathfind + validate move budget for highlight preview only.
 */
export function computeLegalMoveDestinations(
  snap: ViewSnapshot,
  unitId: number,
  movePoints = CLIENT_DEFAULT_MOVE_POINTS,
): Coord[] {
  const unit = snap.units.find((u) => u.id === unitId && u.hp > 0);
  if (!unit) {
    return [];
  }
  const start: Coord = { x: unit.x, y: unit.y };
  const dist = new Map<string, number>();
  const queue: Coord[] = [start];
  dist.set(coordKey(start), 0);
  const destinations: Coord[] = [];

  while (queue.length > 0) {
    const cur = queue.shift()!;
    const curDist = dist.get(coordKey(cur)) ?? 0;
    for (const nb of neighbors(cur)) {
      const k = coordKey(nb);
      if (dist.has(k)) {
        continue;
      }
      if (!isWalkable(snap, nb, unitId)) {
        continue;
      }
      const nd = curDist + 1;
      if (nd > movePoints) {
        continue;
      }
      dist.set(k, nd);
      queue.push(nb);
      if (nb.x !== start.x || nb.y !== start.y) {
        destinations.push(nb);
      }
    }
  }
  return destinations;
}
