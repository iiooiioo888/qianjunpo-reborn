/**
 * Browser Live: POST /v1/tactical/connect → POST /v1/tactical/enter-battle
 * (`docs/janus-http-mirror.md`, #60).
 */

import { resolveAppUrl } from './paths.js';

export const DEFAULT_CONNECT_PATH = 'v1/tactical/connect';
export const DEFAULT_ENTER_BATTLE_PATH = 'v1/tactical/enter-battle';
export const DEFAULT_STEP_LOCKSTEP_PATH = 'v1/tactical/step-lockstep';

/** pkg/lockstep.CommandDelayFrames (3) + 1 frame until delayed commands execute. */
export const TACTICAL_LOCKSTEP_STEPS_AFTER_COMMAND = 4;

/** Janus HTTP wire: 2 = 雙方自動（持久），見 docs/janus-http-mirror.md */
export const TACTICAL_AUTO_COMMAND_MODE_BOTH = 2;

/**
 * 暫定契約（與核心 LaresAuth/Login HTTP 鏡像並行對齊；路徑／欄位以核心合入後為準）。
 * POST JSON `{ "username", "password" }` → `{ "access_token" }`（或 camelCase `accessToken`）。
 * 對應 gRPC：`qianjunpo.lares.v1.LaresAuth/Login`（見 `proto/lares/v1/lares.proto`）。
 */
export const DEFAULT_MINT_PATH = 'v1/lares/login';

export const LIVE_ACCESS_TOKEN_MISSING_MESSAGE =
  '需提供 Lares 簽發 accessToken（?accessToken=… 或同域一鍵 mint）';

const DEFAULT_MINT_USERNAME = 'smoke';
const DEFAULT_MINT_PASSWORD = 'smoke';

export function parseLiveGatewayConfig(params) {
  const connectUrlRaw = params.get('connectUrl') || DEFAULT_CONNECT_PATH;
  const enterBattleUrlRaw = params.get('enterBattleUrl') || DEFAULT_ENTER_BATTLE_PATH;
  const mintUrlRaw =
    params.get('mintUrl') || params.get('loginUrl') || DEFAULT_MINT_PATH;
  const accessToken = params.get('accessToken') ?? '';
  const zoneId = params.get('zoneId') || 'default';
  const shard = Number(params.get('shard') || '0');
  const mintUsername =
    params.get('mintUser') || params.get('username') || DEFAULT_MINT_USERNAME;
  const mintPassword =
    params.get('mintPass') || params.get('password') || DEFAULT_MINT_PASSWORD;
  return {
    connectUrl: resolveAppUrl(connectUrlRaw),
    connectUrlRaw,
    enterBattleUrl: resolveAppUrl(enterBattleUrlRaw),
    enterBattleUrlRaw,
    mintUrl: resolveAppUrl(mintUrlRaw),
    mintUrlRaw,
    mintUsername,
    mintPassword,
    accessToken,
    zoneId,
    shard,
  };
}

export function hasLiveAccessToken(gw) {
  return String(gw.accessToken ?? '').trim().length > 0;
}

function readAccessTokenFromJson(json) {
  if (!json || typeof json !== 'object') {
    return '';
  }
  const token = json.access_token ?? json.accessToken;
  return typeof token === 'string' ? token.trim() : '';
}

/**
 * 同域暫定 mint：POST `gw.mintUrl`（預設 `v1/lares/login`）。
 * @returns {Promise<string>} Lares access token
 */
export async function mintLiveAccessToken(gw) {
  const body = {
    username: gw.mintUsername,
    password: gw.mintPassword,
  };
  const mintRes = await postJson(gw.mintUrl, body);
  if (!mintRes.res.ok) {
    throw new Error(
      `Mint HTTP ${mintRes.res.status}（${gw.mintUrlRaw}）：${mintRes.text || '(empty)'}`,
    );
  }
  const token = readAccessTokenFromJson(mintRes.json);
  if (!token) {
    throw new Error(`Mint 回應缺少 access_token（${gw.mintUrlRaw}）`);
  }
  return token;
}

function readSessionId(json) {
  const id = json.session_id ?? json.sessionId;
  return typeof id === 'string' ? id : '';
}

function readBattleId(json, hint) {
  const id = json.battle_id ?? json.battleId;
  if (typeof id === 'string' && id.length > 0) {
    return id;
  }
  return hint;
}

function extractInitialSnapshot(json) {
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
      return JSON.parse(trimmed);
    } catch {
      return null;
    }
  }
  if (typeof vs === 'object') {
    return vs;
  }
  return null;
}

async function postJson(url, body) {
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  const text = await res.text();
  let json = null;
  if (text) {
    try {
      json = JSON.parse(text);
    } catch {
      json = null;
    }
  }
  return { res, json, text };
}

/**
 * @returns {Promise<{ sessionId: string, battleId: string, initialSnapshot: object | null }>}
 */
export async function prepareLiveJanusSession(gw, battleIdHint = 'default/0') {
  if (!String(gw.accessToken ?? '').trim()) {
    throw new Error(LIVE_ACCESS_TOKEN_MISSING_MESSAGE);
  }
  const connectBody = {
    access_token: gw.accessToken,
    client_version: '0.1.0-shell',
    target_zone: { zone_id: gw.zoneId, shard: gw.shard },
  };
  const connectRes = await postJson(gw.connectUrl, connectBody);
  if (!connectRes.json || !connectRes.res.ok) {
    throw new Error(`Connect HTTP ${connectRes.res.status}: ${connectRes.text || '(empty)'}`);
  }
  let sessionId = readSessionId(connectRes.json);
  if (!sessionId) {
    throw new Error('Connect 回應缺少 session_id');
  }

  const enterRes = await postJson(gw.enterBattleUrl, {
    session_id: sessionId,
    access_token: gw.accessToken,
    target_zone: { zone_id: gw.zoneId, shard: gw.shard },
  });
  if (!enterRes.json || !enterRes.res.ok) {
    throw new Error(`EnterBattle HTTP ${enterRes.res.status}: ${enterRes.text || '(empty)'}`);
  }
  const battleId = readBattleId(enterRes.json, battleIdHint);
  sessionId = readSessionId(enterRes.json) || sessionId;
  const initialSnapshot = extractInitialSnapshot(enterRes.json);
  return { sessionId, battleId, initialSnapshot };
}

function readLockstepFrame(json) {
  if (!json || typeof json !== 'object') {
    return undefined;
  }
  const frame = json.lockstep_frame ?? json.lockstepFrame;
  return typeof frame === 'number' ? frame : undefined;
}

function readStateHash(json) {
  if (!json || typeof json !== 'object') {
    return undefined;
  }
  const hash = json.state_hash ?? json.stateHash;
  return typeof hash === 'number' ? hash : undefined;
}

/**
 * POST step-lockstep (`docs/janus-http-mirror.md`).
 * @returns {Promise<{ ok: boolean, lockstepFrame?: number, stateHash?: number, finished?: boolean, winner?: number, viewSnapshot: object | null, status: number, text: string }>}
 */
export async function stepTacticalLockstep(
  stepUrl,
  {
    sessionId,
    battleId,
    steps = TACTICAL_LOCKSTEP_STEPS_AFTER_COMMAND,
    autoCommandMode,
  } = {},
) {
  const body = {
    session_id: sessionId,
    battle_id: battleId,
    steps,
  };
  if (autoCommandMode != null) {
    body.auto_command_mode = autoCommandMode;
  }
  const stepRes = await postJson(stepUrl, body);
  const json = stepRes.json;
  return {
    ok: stepRes.res.ok && json != null,
    lockstepFrame: readLockstepFrame(json),
    stateHash: readStateHash(json),
    finished: Boolean(json?.finished),
    winner: typeof json?.winner === 'number' ? json.winner : undefined,
    viewSnapshot: extractInitialSnapshot(json),
    status: stepRes.res.status,
    text: stepRes.text,
  };
}
