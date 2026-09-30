import { computeLegalMoveDestinations } from './reachability.js';

/** Display-only mock apply (not authoritative). */
export function applyMockMove(snap, unitId, to) {
  const unit = snap.units.find((u) => u.id === unitId && u.hp > 0);
  if (!unit) {
    return { ok: false, reason: 'unit missing' };
  }
  const legal = computeLegalMoveDestinations(snap, unitId);
  const allowed = legal.some((c) => c.x === to.x && c.y === to.y);
  if (!allowed) {
    return { ok: false, reason: 'destination not in client reachability set' };
  }

  const next = JSON.parse(JSON.stringify(snap));
  const u = next.units.find((x) => x.id === unitId);
  if (!u) {
    return { ok: false, reason: 'unit missing after clone' };
  }
  const fromX = u.x;
  const fromY = u.y;
  next.cells[fromY][fromX].unitId = 0;
  next.cells[to.y][to.x].unitId = unitId;
  u.x = to.x;
  u.y = to.y;
  next.lockstepFrame += 1;
  return { ok: true, snapshot: next };
}
