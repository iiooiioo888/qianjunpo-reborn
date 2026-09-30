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
