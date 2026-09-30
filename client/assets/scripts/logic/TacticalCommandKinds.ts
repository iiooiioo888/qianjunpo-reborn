/** Aligned with pkg/tactical.CommandKind */
export const TacticalCommandKind = {
  Move: 1,
  Attack: 2,
  Pass: 3,
} as const;

export interface SubmitMovePayload {
  battleId: string;
  sessionId: string;
  playerId: number;
  unitId: number;
  toX: number;
  toY: number;
}

export interface SubmitMoveResult {
  accepted: boolean;
  rejectReason?: string;
  stubOnly?: boolean;
  lockstepFrame?: number;
}
