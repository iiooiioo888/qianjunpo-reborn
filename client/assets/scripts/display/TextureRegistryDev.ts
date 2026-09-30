import {
  CHAR_CARD_TEXTURE_KEYS,
  CharacterCardSpriteRegistry,
} from './CharacterCardSpriteRegistry';
import { ICON_ASSET_IDS, IconSpriteRegistry } from './IconSpriteRegistry';
import { UNIT_TEXTURE_KEYS, UnitSpriteRegistry } from './UnitSpriteRegistry';

export interface TexturePreloadReport {
  units: Record<string, boolean>;
  characterCards: Record<string, boolean>;
  hudIcons: Record<string, boolean>;
}

/**
 * Mock／Live 預覽前：在此 `registerResourcePath` 換 v03 或 client 極簡占位 stem，再 `preloadTacticalDisplayTextures()`。
 *
 * 過審後一行換圖範例（取消註解、放入 PNG、重新預覽）：
 * ```ts
 * UnitSpriteRegistry.registerResourcePath(UNIT_TEXTURE_KEYS.infantry, 'PX2D_unit_infantry_v03');
 * CharacterCardSpriteRegistry.registerResourcePath(CHAR_CARD_TEXTURE_KEYS.char_caocao, 'PX2D_CHAR_WEI_Caocao_ex_v03');
 * ```
 *
 * 本地極簡占位（非 art，128×128 / 320×400）檔名範例見 `assets/resources/textures/2d/README.md`。
 */
export function applyDevTextureStemOverrides(): void {
  // UnitSpriteRegistry.registerResourcePath(UNIT_TEXTURE_KEYS.infantry, 'mock_infantry_128');
  // UnitSpriteRegistry.registerResourcePath(UNIT_TEXTURE_KEYS.cavalry, 'mock_cavalry_128');
  // CharacterCardSpriteRegistry.registerResourcePath(CHAR_CARD_TEXTURE_KEYS.char_caocao, 'mock_char_caocao_320x400');
}

const HUD_ICON_IDS = [
  ICON_ASSET_IDS.resFood,
  ICON_ASSET_IDS.resWood,
  ICON_ASSET_IDS.resStone,
  ICON_ASSET_IDS.resIron,
  ICON_ASSET_IDS.resGold,
] as const;

const UNIT_KEYS_FOR_REPORT = [
  UNIT_TEXTURE_KEYS.infantry,
  UNIT_TEXTURE_KEYS.cavalry,
] as const;

const CHAR_KEYS_FOR_REPORT = [
  CHAR_CARD_TEXTURE_KEYS.char_caocao,
  CHAR_CARD_TEXTURE_KEYS.char_zhangfei,
  CHAR_CARD_TEXTURE_KEYS.char_wu_placeholder,
] as const;

export async function preloadTacticalDisplayTextures(): Promise<TexturePreloadReport> {
  applyDevTextureStemOverrides();
  await Promise.all([
    UnitSpriteRegistry.preload(),
    IconSpriteRegistry.preload(),
    CharacterCardSpriteRegistry.preload(),
  ]);
  return buildTexturePreloadReport();
}

export function buildTexturePreloadReport(): TexturePreloadReport {
  const units: Record<string, boolean> = {};
  for (const key of UNIT_KEYS_FOR_REPORT) {
    units[key] = UnitSpriteRegistry.isSpriteLoaded(key);
  }
  const characterCards: Record<string, boolean> = {};
  for (const key of CHAR_KEYS_FOR_REPORT) {
    characterCards[key] = CharacterCardSpriteRegistry.isSpriteLoaded(key);
  }
  const hudIcons: Record<string, boolean> = {};
  for (const id of HUD_ICON_IDS) {
    hudIcons[id] = IconSpriteRegistry.isSpriteLoaded(id);
  }
  return { units, characterCards, hudIcons };
}

/** Mock HUD 第三行附註：缺圖時說明占位，不 crash。 */
export function formatTexturePreloadHudNote(report: TexturePreloadReport): string {
  const missingUnits = Object.entries(report.units).filter(([, ok]) => !ok).map(([k]) => k);
  const missingCards = Object.entries(report.characterCards).filter(([, ok]) => !ok).map(([k]) => k);
  const missingIcons = Object.entries(report.hudIcons).filter(([, ok]) => !ok).length;
  const anyMissing =
    missingUnits.length > 0 || missingCards.length > 0 || missingIcons > 0;
  if (!anyMissing) {
    return '貼圖：registry 已載入（Nearest／整數倍）';
  }
  const parts: string[] = [];
  if (missingUnits.length) {
    parts.push(`單位占位(${missingUnits.join(',')})`);
  }
  if (missingCards.length) {
    parts.push(`角色卡占位(${missingCards.join(',')})`);
  }
  if (missingIcons > 0) {
    parts.push(`資源圖標缺${missingIcons}`);
  }
  return `貼圖：${parts.join('；')} — sync-wip 或 registerResourcePath`;
}

export function logTexturePreloadReport(report: TexturePreloadReport, mode: 'mock' | 'live'): void {
  console.info(`[TextureRegistryDev] ${mode} preload`, report);
}
