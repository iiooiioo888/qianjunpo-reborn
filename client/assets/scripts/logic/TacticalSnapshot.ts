/**
 * Display-layer tactical snapshot (mirrors pkg/tactical ViewSnapshot JSON).
 * No Cocos imports — safe for unit tests / future shared bundle.
 */

export const BOARD_SIZE = 19;

export enum TerrainKind {
  Plain = 0,
  Mountain = 1,
  Forest = 2,
  River = 3,
  City = 4,
  Pass = 5,
}

export interface ViewCell {
  terrain: TerrainKind;
  passable: boolean;
  unitId: number;
}

export interface ViewUnit {
  id: number;
  owner: number;
  type: number;
  x: number;
  y: number;
  hp: number;
  maxHp: number;
}

export interface ViewSnapshot {
  schemaVersion: number;
  boardSize: number;
  seed: number;
  lockstepFrame: number;
  timeFlowRateParts: number;
  initialStateHash: string;
  cells: ViewCell[][];
  units: ViewUnit[];
}

export function parseViewSnapshot(raw: unknown): ViewSnapshot {
  const o = raw as ViewSnapshot;
  if (!o || o.schemaVersion !== 1 || o.boardSize !== BOARD_SIZE) {
    throw new Error('tactical: unsupported or invalid view snapshot schema');
  }
  if (!Array.isArray(o.cells) || o.cells.length !== BOARD_SIZE) {
    throw new Error('tactical: cells grid size mismatch');
  }
  return o;
}

/** Parts-per-10000 rate (10000 = 1.0x), aligned with pkg/timedilation.RateScale. */
export function timeFlowRateToFloat(parts: number): number {
  return parts / 10000;
}

/** Display scale: 10000 parts == 100%. */
export const TIME_FLOW_RATE_SCALE = 10000;

/** Bar full width at catch-up cap (1.5x), aligned with pkg/timedilation.CatchUpCapRate. */
export const TIME_FLOW_BAR_CAP_PARTS = 15000;

/** Lockstep frame → simulated time (ms); no wall clock. */
export const LOCKSTEP_MS_PER_FRAME = 100;

/** Percent string from snapshot parts only (10000 → 100.0%). */
export function formatTimeFlowRatePercent(parts: number): string {
  return `${(parts / 100).toFixed(1)}%`;
}

/** Bar fill ratio in [0, 1], capped at {@link TIME_FLOW_BAR_CAP_PARTS}. */
export function timeFlowBarFillRatio(parts: number): number {
  const capped = Math.min(Math.max(0, parts), TIME_FLOW_BAR_CAP_PARTS);
  return capped / TIME_FLOW_BAR_CAP_PARTS;
}

/** Simulated time label from lockstep frame (frame × 100ms). */
export function formatSimTimeFromLockstepFrame(lockstepFrame: number): string {
  return `sim_time: ${lockstepFrame * LOCKSTEP_MS_PER_FRAME}ms`;
}

/** Overload tint when sim runs slower than 1.0x (parts < 10000). */
export function isTimeFlowOverload(parts: number): boolean {
  return parts < TIME_FLOW_RATE_SCALE;
}
