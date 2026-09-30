import { BOARD_SIZE, CLIENT_DEFAULT_MOVE_POINTS } from './config.js';

const NEIGHBOR_DELTAS = [
  [0, -1],
  [1, -1],
  [1, 0],
  [1, 1],
  [0, 1],
  [-1, 1],
  [-1, 0],
  [-1, -1],
];

function inBounds(x, y) {
  return x >= 0 && x < BOARD_SIZE && y >= 0 && y < BOARD_SIZE;
}

function coordKey(x, y) {
  return `${x},${y}`;
}

function isWalkable(snap, x, y, ignoreUnitId) {
  const cell = snap.cells[y]?.[x];
  if (!cell || !cell.passable) {
    return false;
  }
  if (cell.unitId !== 0 && cell.unitId !== ignoreUnitId) {
    return false;
  }
  return true;
}

/**
 * BFS reachable destinations within movePoints (8-neighbor steps).
 * Display-only; mirrors client/assets ClientMoveReachability.ts.
 */
export function computeLegalMoveDestinations(snap, unitId, movePoints = CLIENT_DEFAULT_MOVE_POINTS) {
  const unit = snap.units.find((u) => u.id === unitId && u.hp > 0);
  if (!unit) {
    return [];
  }
  const startX = unit.x;
  const startY = unit.y;
  const dist = new Map();
  const queue = [{ x: startX, y: startY }];
  dist.set(coordKey(startX, startY), 0);
  const destinations = [];

  while (queue.length > 0) {
    const cur = queue.shift();
    const curDist = dist.get(coordKey(cur.x, cur.y)) ?? 0;
    for (const [dx, dy] of NEIGHBOR_DELTAS) {
      const nx = cur.x + dx;
      const ny = cur.y + dy;
      const k = coordKey(nx, ny);
      if (dist.has(k) || !inBounds(nx, ny)) {
        continue;
      }
      if (!isWalkable(snap, nx, ny, unitId)) {
        continue;
      }
      const nd = curDist + 1;
      if (nd > movePoints) {
        continue;
      }
      dist.set(k, nd);
      queue.push({ x: nx, y: ny });
      if (nx !== startX || ny !== startY) {
        destinations.push({ x: nx, y: ny });
      }
    }
  }
  return destinations;
}
