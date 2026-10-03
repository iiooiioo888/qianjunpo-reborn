/**
 * Live 進入：若目標戰區戰局已終局，改用新 zone_id（仍走 connect → enter-battle）。
 */

import { isBattleFinished } from './battle-end.js';
import { resolveAppUrl } from './paths.js';

/** @param {string} zoneId @param {number} shard */
export function battleIdForZone(zoneId, shard) {
  return `${zoneId}/${shard}`;
}

/**
 * 終局判斷（與 overlay 一致，並涵蓋 step-lockstep `finished`）。
 * @param {object | null | undefined} snap
 */
export function isTerminalLiveSnapshot(snap) {
  if (!snap) {
    return false;
  }
  if (isBattleFinished(snap)) {
    return true;
  }
  if (snap.finished === true) {
    return true;
  }
  const reason = snap.endReason;
  if (reason === 'wipeout' || reason === 'timeout' || reason === 'mutual_wipe' || reason === 'occupy') {
    return true;
  }
  const w = snap.winner;
  return w != null && typeof w === 'number';
}

/** @returns {string} */
export function mintFreshLiveZoneId(prefix = 'qjp') {
  const t = Date.now().toString(36);
  const r = Math.random().toString(36).slice(2, 8);
  return `${prefix}-${t}-${r}`;
}

/**
 * @param {string} liveUrlRaw template from boot (`v1/tactical/snapshot?battle_id=…`)
 * @param {string} battleId
 */
export function liveSnapshotUrlForBattleId(liveUrlRaw, battleId) {
  if (liveUrlRaw.includes('battle_id=')) {
    return resolveAppUrl(
      liveUrlRaw.replace(/battle_id=[^&]+/, `battle_id=${encodeURIComponent(battleId)}`),
    );
  }
  const sep = liveUrlRaw.includes('?') ? '&' : '?';
  return resolveAppUrl(`${liveUrlRaw}${sep}battle_id=${encodeURIComponent(battleId)}`);
}

/**
 * @returns {Promise<object | null>} null when battle missing or unreadable
 */
export async function fetchLiveSnapshotForBattle(boot, battleId) {
  const url = liveSnapshotUrlForBattleId(boot.liveUrlRaw, battleId);
  try {
    const res = await fetch(url, { cache: 'no-store', mode: 'cors' });
    if (res.status === 404) {
      return null;
    }
    if (!res.ok) {
      return null;
    }
    const body = await res.json();
    if (!body || body.schemaVersion !== 1) {
      return null;
    }
    return body;
  } catch {
    return null;
  }
}

/**
 * 將 `zoneId` 寫回 query（不 reload），方便分享／刷新仍進新戰區。
 * @param {string} zoneId
 */
export function persistZoneIdInLocation(zoneId) {
  try {
    const url = new URL(window.location.href);
    url.searchParams.set('zoneId', zoneId);
    window.history.replaceState(null, '', `${url.pathname}${url.search}${url.hash}`);
  } catch {
    // ignore
  }
}

/**
 * 若目前 `boot.liveGateway` 指向的戰區戰局已終局，分配新 zone 並更新 boot／URL。
 *
 * @param {import('./config.js').parseBootConfig extends Function ? ReturnType<import('./config.js').parseBootConfig> : object} boot
 * @returns {Promise<
 *   | { action: 'unchanged'; battleId: string }
 *   | { action: 'rotated'; previousZoneId: string; zoneId: string; battleId: string; endReason: string }
 * >}
 */
export async function ensureFreshLiveZoneBeforeEnter(boot) {
  const gw = boot.liveGateway;
  const shard = Number(gw.shard) || 0;
  const previousZoneId = gw.zoneId || 'default';
  const battleId = battleIdForZone(previousZoneId, shard);
  const snap = await fetchLiveSnapshotForBattle(boot, battleId);
  if (!isTerminalLiveSnapshot(snap)) {
    return { action: 'unchanged', battleId };
  }
  const zoneId = mintFreshLiveZoneId();
  gw.zoneId = zoneId;
  persistZoneIdInLocation(zoneId);
  const newBattleId = battleIdForZone(zoneId, shard);
  return {
    action: 'rotated',
    previousZoneId,
    zoneId,
    battleId: newBattleId,
    endReason: typeof snap?.endReason === 'string' ? snap.endReason : 'terminal',
  };
}
