/**
 * Browser Live: POST /v1/tactical/connect → POST /v1/tactical/enter-battle
 * (`docs/janus-http-mirror.md`, #60).
 */

import { resolveAppUrl } from './paths.js';

export const DEFAULT_CONNECT_PATH = 'v1/tactical/connect';
export const DEFAULT_ENTER_BATTLE_PATH = 'v1/tactical/enter-battle';

export function parseLiveGatewayConfig(params) {
  const connectUrlRaw = params.get('connectUrl') || DEFAULT_CONNECT_PATH;
  const enterBattleUrlRaw = params.get('enterBattleUrl') || DEFAULT_ENTER_BATTLE_PATH;
  const accessToken = params.get('accessToken') || 'dev';
  const zoneId = params.get('zoneId') || 'default';
  const shard = Number(params.get('shard') || '0');
  return {
    connectUrl: resolveAppUrl(connectUrlRaw),
    connectUrlRaw,
    enterBattleUrl: resolveAppUrl(enterBattleUrlRaw),
    enterBattleUrlRaw,
    accessToken,
    zoneId,
    shard,
  };
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
