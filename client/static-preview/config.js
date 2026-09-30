import { resolveAppUrl } from './paths.js';

/** Bundled mock for offline / deploy-web-preview (static-preview only). */
export const MOCK_SNAPSHOT_URL = 'mock/demo_initial.json';
export const MOCK_SNAPSHOT_PATH = MOCK_SNAPSHOT_URL;

/**
 * Same-origin Janus mirror (relative to page base: /qjp/v1/… or :18093 /v1/…).
 * Override with `?liveUrl=` (encodeURIComponent). No :18090/:18092/:18093 in defaults.
 */
export const DEFAULT_LIVE_SNAPSHOT_URL = 'v1/tactical/snapshot?battle_id=default/0';

/** Local player owner id in duel mock (owner 0). */
export const LOCAL_PLAYER_OWNER = 0;

export const BOARD_SIZE = 19;
export const CLIENT_DEFAULT_MOVE_POINTS = 4;

export const LIVE_POLL_INTERVAL_MS = 400;

/** Optional v04 cards — drop files under `chars/` or show placeholder. */
export const CHAR_CARD_V04 = [
  { label: '曹操', src: 'chars/PX2D_CHAR_WEI_Caocao_ex_v04.png' },
  { label: '張飛', src: 'chars/PX2D_CHAR_SHU_Zhangfei_ex_v04.png' },
  { label: '吳', src: 'chars/PX2D_CHAR_WU_Placeholder_01_v04.png' },
];

export function parseBootConfig() {
  const params = new URLSearchParams(window.location.search);
  const live = params.get('live') === '1' || params.get('live') === 'true';
  const liveUrlRaw = params.get('liveUrl') || DEFAULT_LIVE_SNAPSHOT_URL;
  const liveUrl = resolveAppUrl(liveUrlRaw);
  const showCards = params.get('cards') !== '0';
  return { live, liveUrl, liveUrlRaw, showCards };
}

export function mockSnapshotUrl() {
  return resolveAppUrl(MOCK_SNAPSHOT_PATH);
}
