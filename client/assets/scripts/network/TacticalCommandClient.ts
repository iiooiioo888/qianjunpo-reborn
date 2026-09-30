import {
  SubmitMovePayload,
  SubmitMoveResult,
  TacticalCommandKind,
} from '../logic/TacticalCommandKinds';
import { TacticalNetworkConfig } from './JanusGatewayStub';

/**
 * Janus SubmitTacticalCommand (gRPC) — browser preview has no gRPC yet.
 * Optional dev POST mirror (same host as snapshot); if missing, returns stubOnly.
 */
export async function submitTacticalMove(
  cfg: TacticalNetworkConfig,
  payload: SubmitMovePayload,
): Promise<SubmitMoveResult> {
  const url = `http://${cfg.janusHost}:${cfg.janusHttpPort}/v1/tactical/command`;
  const body = {
    session_id: payload.sessionId,
    battle_id: payload.battleId,
    player_id: payload.playerId,
    kind: TacticalCommandKind.Move,
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
      console.info('[TacticalCommandClient] no HTTP command mirror; stub payload', body);
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
    };
    return {
      accepted: Boolean(json.accepted),
      rejectReason: json.reject_reason,
      lockstepFrame: json.lockstep_frame,
    };
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    console.info('[TacticalCommandClient] submit failed; stub payload', body, msg);
    return {
      accepted: false,
      stubOnly: true,
      rejectReason: `網路錯誤（已記錄指令）：${msg}`,
    };
  }
}
