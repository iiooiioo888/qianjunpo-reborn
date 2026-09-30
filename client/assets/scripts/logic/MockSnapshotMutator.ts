import { Coord } from './BoardCoord';
import { computeLegalMoveDestinations } from './ClientMoveReachability';
import { ViewSnapshot } from './TacticalSnapshot';

export interface MockMoveResult {
  ok: boolean;
  reason?: string;
  snapshot?: ViewSnapshot;
}

/** Display-only mock apply (not authoritative). Server / Roma validates on live submit. */
export function applyMockMove(snap: ViewSnapshot, unitId: number, to: Coord): MockMoveResult {
  const unit = snap.units.find((u) => u.id === unitId && u.hp > 0);
  if (!unit) {
    return { ok: false, reason: 'unit missing' };
  }
  const legal = computeLegalMoveDestinations(snap, unitId);
  const allowed = legal.some((c) => c.x === to.x && c.y === to.y);
  if (!allowed) {
    return { ok: false, reason: 'destination not in client reachability set' };
  }

  const next: ViewSnapshot = JSON.parse(JSON.stringify(snap)) as ViewSnapshot;
  const u = next.units.find((x) => x.id === unitId);
  if (!u) {
    return { ok: false, reason: 'unit missing after clone' };
  }
  const from = { x: u.x, y: u.y };
  next.cells[from.y][from.x].unitId = 0;
  next.cells[to.y][to.x].unitId = unitId;
  u.x = to.x;
  u.y = to.y;
  next.lockstepFrame += 1;
  return { ok: true, snapshot: next };
}
