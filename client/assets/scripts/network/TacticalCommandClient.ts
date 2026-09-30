import {
  SubmitMovePayload,
  SubmitMoveResult,
  TacticalCommandKind,
} from '../logic/TacticalCommandKinds';
import { resolveTacticalHttpUrl, TacticalNetworkConfig } from './JanusGatewayStub';

export interface SubmitTacticalCommandPayload extends SubmitMovePayload {
  kind: number;
}

function parseCommandResponse(json: {
  accepted?: boolean;
  reject_reason?: string;
  lockstep_frame?: number;
  state_hash?: number;
}): SubmitMoveResult {
  return {
    accepted: Boolean(json.accepted),
    rejectReason: json.reject_reason,
    lockstepFrame: json.lockstep_frame,
    stateHash: json.state_hash,
  };
}

/**
 * Janus SubmitTacticalCommand — browser uses same-origin POST mirror (#50).
 */
export async function submitTacticalCommand(
  cfg: TacticalNetworkConfig,
  payload: SubmitTacticalCommandPayload,
): Promise<SubmitMoveResult> {
  const url = resolveTacticalHttpUrl(cfg, 'v1/tactical/command');
  const body = {
    session_id: payload.sessionId,
    battle_id: payload.battleId,
    player_id: payload.playerId,
    kind: payload.kind,
    unit_id: payload.unitId,
    to_x: payload.toX,
    to_y: payload.toY,
  };
  try {
    const res = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (res.status === 404 || res.status === 405) {
      console.info('[TacticalCommandClient] no HTTP command mirror; payload', body);
      return {
        accepted: false,
        stubOnly: true,
        rejectReason: 'HTTP /v1/tactical/command 未部署；請用 grpcurl SubmitTacticalCommand',
      };
    }
    if (!res.ok) {
      const text = await res.text();
      return { accepted: false, rejectReason: `Janus command HTTP ${res.status}: ${text}` };
    }
    const json = (await res.json()) as {
      accepted?: boolean;
      reject_reason?: string;
      lockstep_frame?: number;
      state_hash?: number;
    };
    return parseCommandResponse(json);
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    console.warn('[TacticalCommandClient] submit failed', body, msg);
    return {
      accepted: false,
      rejectReason: `網路錯誤：${msg}`,
    };
  }
}

export async function submitTacticalMove(
  cfg: TacticalNetworkConfig,
  payload: SubmitMovePayload,
): Promise<SubmitMoveResult> {
  return submitTacticalCommand(cfg, { ...payload, kind: TacticalCommandKind.Move });
}
