import { resources, Sprite, SpriteFrame, UITransform } from 'cc';
import { applyPixelArtSampling, boardUnitDisplaySize, UNIT_ART_CANVAS_PX } from './PixelSpriteUtil';

/**
 * 邏輯鍵（穩定 API）；過審後僅改 {@link registerResourcePath} 或下方預設 stem 即可換 v03 檔名。
 * 與 art/2d/_wip/units 檔名（無擴展名）對齊；PNG 放在 `resources/textures/2d/units/`。
 */
export const UNIT_TEXTURE_KEYS = {
  infantry: 'infantry',
  cavalry: 'cavalry',
} as const;

export type UnitTextureKey = (typeof UNIT_TEXTURE_KEYS)[keyof typeof UNIT_TEXTURE_KEYS];

/** 預設資源 stem（無路徑、無擴展名）；可被 registerResourcePath 覆寫。 */
const DEFAULT_STEM_BY_KEY: Record<UnitTextureKey, string> = {
  infantry: 'PX2D_unit_infantry_v02',
  cavalry: 'PX2D_unit_cavalry_v02',
};

/** ViewUnit.type → registry 鍵（與 pkg/tactical 單位種類對齊，缺省回步兵）。 */
const TYPE_TO_KEY: Record<number, UnitTextureKey> = {
  0: UNIT_TEXTURE_KEYS.infantry,
  2: UNIT_TEXTURE_KEYS.cavalry,
};

const DEFAULT_TYPE = 0;

function unitsResourcePath(stem: string): string {
  return `textures/2d/units/${stem}`;
}

/**
 * Lazy-loads PX2D unit sprites from resources/textures/2d/units.
 * Missing assets resolve to null; {@link UnitPlaceholderView} 退回 Graphics — no throw.
 */
export class UnitSpriteRegistry {
  private static stems = new Map<string, string>(Object.entries(DEFAULT_STEM_BY_KEY));
  private static frames = new Map<string, SpriteFrame | null>();
  private static loadPromise: Promise<void> | null = null;

  /** 換圖：例如 v03 過審後 `registerResourcePath('infantry', 'PX2D_unit_infantry_v03')` 再 `preload()`。 */
  static registerResourcePath(key: UnitTextureKey | string, resourceStem: string): void {
    this.stems.set(key, resourceStem);
    this.frames.delete(key);
    this.loadPromise = null;
  }

  static preload(): Promise<void> {
    if (this.loadPromise) {
      return this.loadPromise;
    }
    const keys = [...new Set([...this.stems.keys(), ...Object.values(TYPE_TO_KEY)])];
    this.loadPromise = Promise.all(keys.map((key) => this.loadKey(key))).then(() => undefined);
    return this.loadPromise;
  }

  private static loadKey(key: string): Promise<void> {
    const stem = this.stems.get(key);
    if (!stem) {
      this.frames.set(key, null);
      return Promise.resolve();
    }
    const path = unitsResourcePath(stem);
    return new Promise((resolve) => {
      resources.load(path, SpriteFrame, (err, sf) => {
        if (err || !sf) {
          console.warn(`[UnitSpriteRegistry] sprite missing: ${path}`, err?.message ?? '');
          this.frames.set(key, null);
        } else {
          applyPixelArtSampling(sf);
          this.frames.set(key, sf);
        }
        resolve();
      });
    });
  }

  static resolveKeyForUnitType(unitType: number): UnitTextureKey {
    return TYPE_TO_KEY[unitType] ?? TYPE_TO_KEY[DEFAULT_TYPE];
  }

  static getSpriteFrameByKey(key: UnitTextureKey | string): SpriteFrame | null {
    const cached = this.frames.get(key);
    if (cached !== undefined) {
      return cached;
    }
    return null;
  }

  /** @deprecated 請優先 {@link getSpriteFrameByKey}；仍依 type 映射鍵。 */
  static getSpriteFrame(unitType: number): SpriteFrame | null {
    const key = this.resolveKeyForUnitType(unitType);
    const direct = this.getSpriteFrameByKey(key);
    if (direct) {
      return direct;
    }
    const fallbackKey = TYPE_TO_KEY[DEFAULT_TYPE];
    return this.getSpriteFrameByKey(fallbackKey);
  }

  /**
   * 套用單位貼圖到 Sprite；缺圖時關閉 Sprite（由占位 Graphics 接手）。
   * @returns 是否成功套用
   */
  static applyUnitToSprite(
    sprite: Sprite | null,
    unitType: number,
    cellSizePx: number,
    artCanvasPx = UNIT_ART_CANVAS_PX,
  ): boolean {
    if (!sprite) {
      return false;
    }
    const sf = this.getSpriteFrame(unitType);
    if (!sf) {
      sprite.enabled = false;
      sprite.spriteFrame = null;
      return false;
    }
    const display = boardUnitDisplaySize(artCanvasPx, cellSizePx);
    sprite.enabled = true;
    sprite.sizeMode = Sprite.SizeMode.CUSTOM;
    sprite.spriteFrame = sf;
    const ui = sprite.node.getComponent(UITransform) ?? sprite.node.addComponent(UITransform);
    ui.setContentSize(display, display);
    return true;
  }

  static hasAnySprite(): boolean {
    for (const sf of this.frames.values()) {
      if (sf) {
        return true;
      }
    }
    return false;
  }

  /** preload 後：有 SpriteFrame 為 true；缺圖登記為 false（未 preload 亦 false）。 */
  static isSpriteLoaded(key: UnitTextureKey | string): boolean {
    const sf = this.frames.get(key);
    return sf != null;
  }
}
