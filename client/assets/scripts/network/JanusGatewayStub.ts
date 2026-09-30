/**
 * Janus → Roma network boundary. Authoritative state stays on Roma;
 * client display consumes view snapshots only.
 */
export interface TacticalNetworkConfig {
  janusHost: string;
  janusGrpcPort: number;
  janusHttpPort: number;
  clientVersion: string;
}

export const DEFAULT_NETWORK_STUB: TacticalNetworkConfig = {
  janusHost: '127.0.0.1',
  janusGrpcPort: 9090,
  janusHttpPort: 8090,
  clientVersion: '0.1.0-shell',
};

/** Dev HTTP mirror of JanusGateway.GetBattleSnapshot (browser-friendly). */
export async function fetchLiveViewSnapshot(
  cfg: TacticalNetworkConfig,
  battleId: string,
  sessionId = '',
): Promise<unknown> {
  const q = new URLSearchParams({ battle_id: battleId });
  if (sessionId) {
    q.set('session_id', sessionId);
  }
  const url = `http://${cfg.janusHost}:${cfg.janusHttpPort}/v1/tactical/snapshot?${q.toString()}`;
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`Janus snapshot HTTP ${res.status}: ${await res.text()}`);
  }
  return res.json();
}

/**
 * gRPC JanusGateway (EnterBattle / StepTacticalLockstep / GetBattleSnapshot) is the
 * production path; use grpcurl or a native gRPC client from tools/CI — not wired in Cocos yet.
 */
export class JanusGatewayStub {
  async connect(_cfg: TacticalNetworkConfig): Promise<{ sessionId: string }> {
    return Promise.reject(new Error('JanusGatewayStub: use fetchLiveViewSnapshot or grpcurl for live snapshots'));
  }
}
