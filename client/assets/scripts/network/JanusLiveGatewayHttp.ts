import { resolveTacticalHttpUrl, TacticalNetworkConfig } from './JanusGatewayStub';

/** Aligned with `docs/janus-http-mirror.md` (#60). */
export const DEFAULT_JANUS_HTTP_CONNECT_PATH = 'v1/tactical/connect';
export const DEFAULT_JANUS_HTTP_ENTER_BATTLE_PATH = 'v1/tactical/enter-battle';

export interface JanusLiveZoneRef {
  zoneId: string;
  shard: number;
}

export interface JanusLiveSessionPrepareOptions {
  cfg: TacticalNetworkConfig;
  accessToken: string;
  clientVersion: string;
  targetZone: JanusLiveZoneRef;
  /** Fallback when EnterBattle omits battle_id (inspector hint). */
  battleIdHint?: string;
}

export interface JanusLiveSessionPrepareResult {
  sessionId: string;
  battleId: string;
  initialSnapshot: unknown | null;
}

export class JanusLiveSessionPrepareError extends Error {
  readonly httpStatus?: number;

  constructor(message: string, httpStatus?: number) {
    super(message);
    this.httpStatus = httpStatus;
  }
}

function connectPath(cfg: TacticalNetworkConfig): string {
  const override = cfg.janusHttpConnectPath?.trim();
  return override || DEFAULT_JANUS_HTTP_CONNECT_PATH;
}

function enterBattlePath(cfg: TacticalNetworkConfig): string {
  const override = cfg.janusHttpEnterBattlePath?.trim();
  return override || DEFAULT_JANUS_HTTP_ENTER_BATTLE_PATH;
}

function zoneBody(zone: JanusLiveZoneRef): { zone_id: string; shard: number } {
  return { zone_id: zone.zoneId, shard: zone.shard };
}

function readSessionId(json: Record<string, unknown>): string {
  const id = json.session_id ?? json.sessionId;
  return typeof id === 'string' ? id : '';
}

function readBattleId(json: Record<string, unknown>, hint: string): string {
  const id = json.battle_id ?? json.battleId;
  if (typeof id === 'string' && id.length > 0) {
    return id;
  }
  return hint;
}

export function extractViewSnapshotFromEnterPayload(json: Record<string, unknown>): unknown | null {
  const vs = json.view_snapshot_json ?? json.viewSnapshotJson;
  if (vs == null) {
    return null;
  }
  if (typeof vs === 'string') {
    const trimmed = vs.trim();
    if (!trimmed) {
      return null;
    }
    try {
      return JSON.parse(trimmed) as unknown;
    } catch {
      return null;
    }
  }
  if (typeof vs === 'object') {
    return vs;
  }
  return null;
}

async function postJson(
  cfg: TacticalNetworkConfig,
  relativePath: string,
  body: Record<string, unknown>,
): Promise<{ status: number; json: Record<string, unknown> | null; text: string }> {
  const url = resolveTacticalHttpUrl(cfg, relativePath);
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  const text = await res.text();
  let json: Record<string, unknown> | null = null;
  if (text) {
    try {
      json = JSON.parse(text) as Record<string, unknown>;
    } catch {
      json = null;
    }
  }
  return { status: res.status, json, text };
}

export async function httpJanusConnect(
  cfg: TacticalNetworkConfig,
  opts: Omit<JanusLiveSessionPrepareOptions, 'battleIdHint'>,
): Promise<string> {
  const path = connectPath(cfg);
  const body: Record<string, unknown> = {
    access_token: opts.accessToken,
    target_zone: zoneBody(opts.targetZone),
  };
  if (opts.clientVersion) {
    body.client_version = opts.clientVersion;
  }
  const { status, json, text } = await postJson(cfg, path, body);
  if (!json || status < 200 || status >= 300) {
    throw new JanusLiveSessionPrepareError(
      `Connect HTTP ${status}: ${text || '(empty)'}`,
      status,
    );
  }
  const sessionId = readSessionId(json);
  if (!sessionId) {
    throw new JanusLiveSessionPrepareError('Connect 回應缺少 session_id');
  }
  return sessionId;
}

export async function httpJanusEnterBattle(
  cfg: TacticalNetworkConfig,
  opts: JanusLiveSessionPrepareOptions & { sessionId: string },
): Promise<JanusLiveSessionPrepareResult> {
  const path = enterBattlePath(cfg);
  const { status, json, text } = await postJson(cfg, path, {
    session_id: opts.sessionId,
    access_token: opts.accessToken,
    target_zone: zoneBody(opts.targetZone),
  });
  if (!json || status < 200 || status >= 300) {
    throw new JanusLiveSessionPrepareError(
      `EnterBattle HTTP ${status}: ${text || '(empty)'}`,
      status,
    );
  }
  const battleId = readBattleId(json, opts.battleIdHint ?? 'default/0');
  if (!battleId) {
    throw new JanusLiveSessionPrepareError('EnterBattle 回應缺少 battle_id');
  }
  const sessionId = readSessionId(json) || opts.sessionId;
  const initialSnapshot = extractViewSnapshotFromEnterPayload(json);
  return { sessionId, battleId, initialSnapshot };
}

/** Live browser: POST connect → POST enter-battle (docs/janus-http-mirror.md). */
export async function prepareJanusLiveSession(
  opts: JanusLiveSessionPrepareOptions,
): Promise<JanusLiveSessionPrepareResult> {
  const sessionId = await httpJanusConnect(opts.cfg, opts);
  return httpJanusEnterBattle(opts.cfg, { ...opts, sessionId });
}

export function formatJanusLivePrepareError(err: unknown): string {
  if (err instanceof JanusLiveSessionPrepareError) {
    return err.message;
  }
  const msg = err instanceof Error ? err.message : String(err);
  return `Live 建局失敗：${msg}`;
}
