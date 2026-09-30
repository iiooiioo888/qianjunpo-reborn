import { resources, Sprite, SpriteFrame, UITransform } from 'cc';
import {
  applyPixelArtSampling,
  CHAR_CARD_ART_HEIGHT_PX,
  CHAR_CARD_ART_WIDTH_PX,
  charCardDisplaySize,
} from './PixelSpriteUtil';

/**
 * 角色卡邏輯鍵；PNG 放在 `resources/textures/2d/chars/`（v03 前可僅登記、不匯入檔案）。
 * 檔名 stem 與 art/2d/_wip/characters 一致，過審後改 stem 或 registerResourcePath 即可。
 */
export const CHAR_CARD_TEXTURE_KEYS = {
  char_caocao: 'char_caocao',
  char_zhangfei: 'char_zhangfei',
  char_wu_placeholder: 'char_wu_placeholder',
} as const;

export type CharCardTextureKey = (typeof CHAR_CARD_TEXTURE_KEYS)[keyof typeof CHAR_CARD_TEXTURE_KEYS];

const DEFAULT_STEM_BY_KEY: Record<CharCardTextureKey, string> = {
  char_caocao: 'PX2D_CHAR_WEI_Caocao_ex_v02',
  char_zhangfei: 'PX2D_CHAR_SHU_Zhangfei_ex_v02',
  char_wu_placeholder: 'PX2D_CHAR_WU_Placeholder_01_v02',
};

function charsResourcePath(stem: string): string {
  return `textures/2d/chars/${stem}`;
}

/**
 * Lazy-loads character card sprites from resources/textures/2d/chars.
 * Missing assets → null；HUD 用占位色塊，不 throw。
 */
export class CharacterCardSpriteRegistry {
  private static stems = new Map<string, string>(Object.entries(DEFAULT_STEM_BY_KEY));
  private static frames = new Map<string, SpriteFrame | null>();
  private static loadPromise: Promise<void> | null = null;

  static registerResourcePath(key: CharCardTextureKey | string, resourceStem: string): void {
    this.stems.set(key, resourceStem);
    this.frames.delete(key);
    this.loadPromise = null;
  }

  static preload(): Promise<void> {
    if (this.loadPromise) {
      return this.loadPromise;
    }
    const keys = [...this.stems.keys()];
    this.loadPromise = Promise.all(keys.map((key) => this.loadKey(key))).then(() => undefined);
    return this.loadPromise;
  }

  private static loadKey(key: string): Promise<void> {
    const stem = this.stems.get(key);
    if (!stem) {
      this.frames.set(key, null);
      return Promise.resolve();
    }
    const path = charsResourcePath(stem);
    return new Promise((resolve) => {
      resources.load(path, SpriteFrame, (err, sf) => {
        if (err || !sf) {
          console.warn(`[CharacterCardSpriteRegistry] card missing: ${path}`, err?.message ?? '');
          this.frames.set(key, null);
        } else {
          applyPixelArtSampling(sf);
          this.frames.set(key, sf);
        }
        resolve();
      });
    });
  }

  static getSpriteFrame(key: CharCardTextureKey | string): SpriteFrame | null {
    const cached = this.frames.get(key);
    if (cached !== undefined) {
      return cached;
    }
    return null;
  }

  /**
   * 將角色卡套到 Sprite；缺圖時關閉 Sprite。
   * 顯示尺寸依 320×400 畫布整數倍縮入 maxWidth×maxHeight。
   */
  static applyCardToSprite(
    sprite: Sprite | null,
    key: CharCardTextureKey | string,
    maxWidth: number,
    maxHeight: number,
  ): boolean {
    if (!sprite) {
      return false;
    }
    const sf = this.getSpriteFrame(key);
    if (!sf) {
      sprite.enabled = false;
      sprite.spriteFrame = null;
      return false;
    }
    const { width, height } = charCardDisplaySize(maxWidth, maxHeight);
    sprite.enabled = true;
    sprite.sizeMode = Sprite.SizeMode.CUSTOM;
    sprite.spriteFrame = sf;
    const ui = sprite.node.getComponent(UITransform) ?? sprite.node.addComponent(UITransform);
    ui.setContentSize(width, height);
    return true;
  }

  static hasAnyCard(): boolean {
    for (const sf of this.frames.values()) {
      if (sf) {
        return true;
      }
    }
    return false;
  }

  static isSpriteLoaded(key: CharCardTextureKey | string): boolean {
    const sf = this.frames.get(key);
    return sf != null;
  }

  /** 文件／除錯用：目標畫布像素。 */
  static artCanvasSize(): { width: number; height: number } {
    return { width: CHAR_CARD_ART_WIDTH_PX, height: CHAR_CARD_ART_HEIGHT_PX };
  }
}
