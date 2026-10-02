import { resources, SpriteFrame } from 'cc';
import { TerrainKind } from '../logic/TacticalSnapshot';
import { applyPixelArtSampling } from './PixelSpriteUtil';

/**
 * ISO25 地格邏輯鍵；PNG stem 對齊 art/25d/_wip/tiles（無擴展名）。
 * 資源路徑：`resources/textures/2d/tiles/`。
 */
export const TERRAIN_TILE_TEXTURE_KEYS = {
  plain: 'plain',
  mountain: 'mountain',
  forest: 'forest',
  river: 'river',
} as const;

export type TerrainTileTextureKey =
  (typeof TERRAIN_TILE_TEXTURE_KEYS)[keyof typeof TERRAIN_TILE_TEXTURE_KEYS];

/** 預設 stem（STANDARD #83）；可被 registerResourcePath 覆寫。 */
const DEFAULT_STEM_BY_KEY: Record<TerrainTileTextureKey, string> = {
  plain: 'ISO25_tile_grass_v02',
  mountain: 'ISO25_tile_mountain_v01',
  forest: 'ISO25_tile_forest_v01',
  river: 'ISO25_tile_water_v01',
};

const TERRAIN_TO_KEY: Partial<Record<TerrainKind, TerrainTileTextureKey>> = {
  [TerrainKind.Plain]: TERRAIN_TILE_TEXTURE_KEYS.plain,
  [TerrainKind.Mountain]: TERRAIN_TILE_TEXTURE_KEYS.mountain,
  [TerrainKind.Forest]: TERRAIN_TILE_TEXTURE_KEYS.forest,
  [TerrainKind.River]: TERRAIN_TILE_TEXTURE_KEYS.river,
};

const DEFAULT_TERRAIN_KEY = TERRAIN_TILE_TEXTURE_KEYS.plain;

function tilesResourcePath(stem: string): string {
  return `textures/2d/tiles/${stem}`;
}

/**
 * Lazy-loads ISO25 terrain tiles from resources/textures/2d/tiles.
 * Missing assets resolve to null; {@link TacticalBoardView} 退回 flat fill — no throw.
 */
export class TerrainTileSpriteRegistry {
  private static stems = new Map<string, string>(Object.entries(DEFAULT_STEM_BY_KEY));
  private static frames = new Map<string, SpriteFrame | null>();
  private static loadPromise: Promise<void> | null = null;

  static registerResourcePath(key: TerrainTileTextureKey | string, resourceStem: string): void {
    this.stems.set(key, resourceStem);
    this.frames.delete(key);
    this.loadPromise = null;
  }

  static preload(): Promise<void> {
    if (this.loadPromise) {
      return this.loadPromise;
    }
    const keys = [...new Set([...this.stems.keys(), ...Object.values(TERRAIN_TO_KEY)])];
    this.loadPromise = Promise.all(keys.map((key) => this.loadKey(key))).then(() => undefined);
    return this.loadPromise;
  }

  private static loadKey(key: string): Promise<void> {
    const stem = this.stems.get(key);
    if (!stem) {
      this.frames.set(key, null);
      return Promise.resolve();
    }
    const path = tilesResourcePath(stem);
    return new Promise((resolve) => {
      resources.load(path, SpriteFrame, (err, sf) => {
        if (err || !sf) {
          console.warn(`[TerrainTileSpriteRegistry] sprite missing: ${path}`, err?.message ?? '');
          this.frames.set(key, null);
        } else {
          applyPixelArtSampling(sf);
          this.frames.set(key, sf);
        }
        resolve();
      });
    });
  }

  static resolveKeyForTerrain(terrain: TerrainKind): TerrainTileTextureKey {
    return TERRAIN_TO_KEY[terrain] ?? DEFAULT_TERRAIN_KEY;
  }

  static getSpriteFrameByKey(key: TerrainTileTextureKey | string): SpriteFrame | null {
    const cached = this.frames.get(key);
    if (cached !== undefined) {
      return cached;
    }
    return null;
  }

  static getSpriteFrameForTerrain(terrain: TerrainKind): SpriteFrame | null {
    const key = this.resolveKeyForTerrain(terrain);
    const direct = this.getSpriteFrameByKey(key);
    if (direct) {
      return direct;
    }
    return this.getSpriteFrameByKey(DEFAULT_TERRAIN_KEY);
  }

  static isSpriteLoaded(key: TerrainTileTextureKey | string): boolean {
    const sf = this.frames.get(key);
    return sf != null;
  }
}
