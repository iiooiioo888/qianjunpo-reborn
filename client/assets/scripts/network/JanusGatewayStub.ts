/**
 * Janus → Roma network boundary (stub). Authoritative state stays on Roma;
 * client display consumes view snapshots only.
 */
export interface TacticalNetworkConfig {
  janusHost: string;
  janusPort: number;
  clientVersion: string;
}

export const DEFAULT_NETWORK_STUB: TacticalNetworkConfig = {
  janusHost: '127.0.0.1',
  janusPort: 7443,
  clientVersion: '0.1.0-shell',
};

/** Placeholder until proto/gateway EnterBattle + StepTacticalLockstep client is wired. */
export class JanusGatewayStub {
  async connect(_cfg: TacticalNetworkConfig): Promise<{ sessionId: string }> {
    return Promise.reject(new Error('JanusGatewayStub: live TCP not implemented in shell PR'));
  }
}
