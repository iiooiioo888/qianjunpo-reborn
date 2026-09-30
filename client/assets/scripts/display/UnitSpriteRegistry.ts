import { resources, SpriteFrame } from 'cc';
import { applyPixelArtSampling } from './PixelSpriteUtil';

/** Tactical unit type → resources path (no extension). Mirrors art/2d/_wip/units v02 IDs. */
const TYPE_TO_RESOURCE: Record<number, string> = {
  0: 'textures/2d/units/PX2D_unit_infantry_v02',
  2: 'textures/2d/units/PX2D_unit_cavalry_v02',
};

const DEFAULT_TYPE = 0;

/**
 * Lazy-loads PX2D unit sprites from resources/textures/2d/units.
 * Missing assets fall back to Graphics placeholders in UnitPlaceholderView.
 */
export class UnitSpriteRegistry {
  private static frames = new Map<number, SpriteFrame | null>();
  private static loadPromise: Promise<void> | null = null;

  static preload(): Promise<void> {
    if (this.loadPromise) {
      return this.loadPromise;
    }
    const types = Object.keys(TYPE_TO_RESOURCE).map((k) => Number(k));
    this.loadPromise = Promise.all(
      types.map(
        (type) =>
          new Promise<void>((resolve) => {
            const path = TYPE_TO_RESOURCE[type];
            resources.load(path, SpriteFrame, (err, sf) => {
              if (err || !sf) {
                console.warn(`[UnitSpriteRegistry] sprite missing: ${path}`, err?.message ?? '');
                this.frames.set(type, null);
              } else {
                applyPixelArtSampling(sf);
                this.frames.set(type, sf);
              }
              resolve();
            });
          }),
      ),
    ).then(() => undefined);
    return this.loadPromise;
  }

  static getSpriteFrame(unitType: number): SpriteFrame | null {
    const direct = this.frames.get(unitType);
    if (direct !== undefined) {
      return direct;
    }
    return this.frames.get(DEFAULT_TYPE) ?? null;
  }

  static hasAnySprite(): boolean {
    for (const sf of this.frames.values()) {
      if (sf) {
        return true;
      }
    }
    return false;
  }
}
