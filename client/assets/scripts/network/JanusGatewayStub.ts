/**
 * Janus → Roma network boundary. Authoritative state stays on Roma;
 * client display consumes view snapshots only.
 */
export interface TacticalNetworkConfig {
  janusHost: string;
  janusGrpcPort: number;
  janusHttpPort: number;
  clientVersion: string;
  /**
   * Optional override for Janus tactical HTTP mirror base.
   * Empty / unset → same-origin relative `v1/tactical/…` (Dev `/qjp/v1/…` or `:18093/v1/…`).
   * Example override: `http://127.0.0.1:8090` for curl-aligned local smoke tests.
   */
  janusHttpTacticalBase?: string;
}

export const DEFAULT_NETWORK_STUB: TacticalNetworkConfig = {
  janusHost: '127.0.0.1',
  janusGrpcPort: 9090,
  janusHttpPort: 8090,
  clientVersion: '0.1.0-shell',
};

/** Resolve Janus tactical HTTP path (same-origin relative by default). */
export function resolveTacticalHttpUrl(cfg: TacticalNetworkConfig, pathAndQuery: string): string {
  const path = pathAndQuery.startsWith('/') ? pathAndQuery.slice(1) : pathAndQuery;
  const base = cfg.janusHttpTacticalBase?.trim();
  if (base) {
    return `${base.replace(/\/$/, '')}/${path}`;
  }
  return path;
}

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
  const url = resolveTacticalHttpUrl(cfg, `v1/tactical/snapshot?${q.toString()}`);
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
