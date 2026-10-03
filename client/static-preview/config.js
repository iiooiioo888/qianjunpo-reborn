import { resolveAppUrl } from './paths.js';
import { parseLiveGatewayConfig } from './live-gateway.js';
import { STUB_SKILL_STRIKE_ID } from './stub-skill.js';
export { TERRAIN_TILE_SRC } from './asset-registry.js';
export { CHAR_CARD_V04 } from './asset-registry.js';

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
export const LIVE_POLL_INTERVAL_MS = 400;

/** Live auto-play: infer suggest → command → step-lockstep (disable with ?auto=0). */
export const LIVE_AUTO_TICK_MS = 1400;

/** POST /v1/suggest + GET write-back (Janus may not mirror; local AI fallback). */
export const DEFAULT_LIVE_SUGGEST_URL = 'v1/suggest';
export const DEFAULT_LIVE_SUGGEST_WRITE_BACK_URL = 'v1/suggest/write-back';

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
  const mockVictoryRaw = (params.get('mockVictory') || params.get('mockEnd') || '').toLowerCase();
  const mockVictory =
    mockVictoryRaw === 'win' || mockVictoryRaw === 'lose' || mockVictoryRaw === 'draw'
      ? mockVictoryRaw
      : '';
  const autoParam = params.get('auto');
  const liveAuto =
    live && autoParam !== '0' && autoParam !== 'false' && autoParam !== 'off';
  const suggestUrlRaw = params.get('suggestUrl') || DEFAULT_LIVE_SUGGEST_URL;
  const suggestWriteBackUrlRaw =
    params.get('suggestWriteBackUrl') || DEFAULT_LIVE_SUGGEST_WRITE_BACK_URL;
  const debugMoves =
    params.get('debug') === '1' ||
    params.get('debugMoves') === '1' ||
    params.get('debug') === 'true';
  const skipCampaign =
    params.get('skipCampaign') === '1' || params.get('skipCampaign') === 'true';
  return {
    live,
    liveAuto,
    debugMoves,
    skipCampaign,
    suggestUrl: resolveAppUrl(suggestUrlRaw),
    suggestUrlRaw,
    suggestWriteBackUrl: resolveAppUrl(suggestWriteBackUrlRaw),
    suggestWriteBackUrlRaw,
    liveUrl,
    liveUrlRaw,
    commandUrl,
    commandUrlRaw,
    stepLockstepUrl,
    stepLockstepUrlRaw,
    showCards,
    liveGateway,
    stubSkillId: Number.isFinite(stubSkillId) && stubSkillId > 0 ? stubSkillId : STUB_SKILL_STRIKE_ID,
    mockVictory,
  };
}

export function mockSnapshotUrl() {
  return resolveAppUrl(MOCK_SNAPSHOT_PATH);
}
