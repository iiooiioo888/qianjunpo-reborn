/** @typedef {import('./types.js').ViewSnapshot} ViewSnapshot */

/** Default mock snapshot (serve from `client/` root). */
export const MOCK_SNAPSHOT_URL = '../assets/resources/data/tactical/demo_initial.json';

/**
 * Dev Janus HTTP mirror (may block cross-origin without CORS / reverse proxy).
 * Override with `?liveUrl=` (encodeURIComponent).
 */
export const DEFAULT_LIVE_SNAPSHOT_URL =
  'http://47.79.23.223:18090/v1/tactical/snapshot?battle_id=default/0';

/** Local player owner id in duel mock (owner 0). */
export const LOCAL_PLAYER_OWNER = 0;

export const BOARD_SIZE = 19;
export const CLIENT_DEFAULT_MOVE_POINTS = 4;

export const LIVE_POLL_INTERVAL_MS = 400;

export const CHAR_CARD_V04 = [
  {
    label: '曹操',
    src: '../assets/resources/textures/2d/chars/PX2D_CHAR_WEI_Caocao_ex_v04.png',
  },
  {
    label: '張飛',
    src: '../assets/resources/textures/2d/chars/PX2D_CHAR_SHU_Zhangfei_ex_v04.png',
  },
  {
    label: '吳',
    src: '../assets/resources/textures/2d/chars/PX2D_CHAR_WU_Placeholder_01_v04.png',
  },
];

export function parseBootConfig() {
  const params = new URLSearchParams(window.location.search);
  const live = params.get('live') === '1' || params.get('live') === 'true';
  const liveUrl = params.get('liveUrl') || DEFAULT_LIVE_SNAPSHOT_URL;
  const showCards = params.get('cards') !== '0';
  return { live, liveUrl, showCards };
}
