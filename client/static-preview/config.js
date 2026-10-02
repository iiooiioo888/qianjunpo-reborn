import { resolveAppUrl } from './paths.js';
import { parseLiveGatewayConfig } from './live-gateway.js';
import { STUB_SKILL_STRIKE_ID } from './stub-skill.js';

/** Bundled mock for offline / deploy-web-preview (static-preview only). */
export const MOCK_SNAPSHOT_URL = 'mock/demo_initial.json';
export const MOCK_SNAPSHOT_PATH = MOCK_SNAPSHOT_URL;

/**
 * Same-origin Janus mirror (relative to page base: /qjp/v1/… or :18093 /v1/…).
 * Override with `?liveUrl=` (encodeURIComponent). No :18090/:18092/:18093 in defaults.
 */
export const DEFAULT_LIVE_SNAPSHOT_URL = 'v1/tactical/snapshot?battle_id=default/0';

/** Same-origin POST mirror (#50); override with `?commandUrl=`. */
export const DEFAULT_LIVE_COMMAND_URL = 'v1/tactical/command';

/** Step Roma lockstep after accepted command; override with `?stepLockstepUrl=`. */
export const DEFAULT_LIVE_STEP_LOCKSTEP_URL = 'v1/tactical/step-lockstep';

/** Local player owner id in duel mock (owner 0). */
export const LOCAL_PLAYER_OWNER = 0;

export const BOARD_SIZE = 19;
export const CLIENT_DEFAULT_MOVE_POINTS = 4;

/** ISO25 STANDARD 地格（art #83）；檔名 stem 與 Cocos TerrainTileSpriteRegistry 一致。 */
export const ISO25_TILE_ART_W = 64;
export const ISO25_TILE_ART_H = 32;
export const TERRAIN_TILE_SRC = {
  0: 'tiles/ISO25_tile_grass_v02.png',
  1: 'tiles/ISO25_tile_mountain_v01.png',
};

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
  const commandUrlRaw = params.get('commandUrl') || DEFAULT_LIVE_COMMAND_URL;
  const commandUrl = resolveAppUrl(commandUrlRaw);
  const stepLockstepUrlRaw = params.get('stepLockstepUrl') || DEFAULT_LIVE_STEP_LOCKSTEP_URL;
  const stepLockstepUrl = resolveAppUrl(stepLockstepUrlRaw);
  const showCards = params.get('cards') !== '0';
  const liveGateway = parseLiveGatewayConfig(params);
  const stubSkillIdRaw = params.get('skillId') ?? params.get('skill_id');
  const stubSkillId = stubSkillIdRaw != null ? Number(stubSkillIdRaw) : STUB_SKILL_STRIKE_ID;
  return {
    live,
    liveUrl,
    liveUrlRaw,
    commandUrl,
    commandUrlRaw,
    stepLockstepUrl,
    stepLockstepUrlRaw,
    showCards,
    liveGateway,
    stubSkillId: Number.isFinite(stubSkillId) && stubSkillId > 0 ? stubSkillId : STUB_SKILL_STRIKE_ID,
  };
}

export function mockSnapshotUrl() {
  return resolveAppUrl(MOCK_SNAPSHOT_PATH);
}
