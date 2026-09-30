import { resources, Sprite, SpriteFrame, UITransform } from 'cc';
import { applyPixelArtSampling } from './PixelSpriteUtil';

/**
 * 與 art/2d/_wip/icons 檔名（無擴展名）一致；過審前僅同步至 client/resources，不寫入 art/2d/icons/。
 * UI 透過 `IconSpriteRegistry.getSpriteFrame(assetId)` 或 `applyIconToSprite` 載入。
 */
export const ICON_ASSET_IDS = {
  army: 'PX2D_icon_army_32',
  battleReport: 'PX2D_icon_battle_report_32',
  bldAcademy: 'PX2D_icon_bld_academy_32',
  bldBarracks: 'PX2D_icon_bld_barracks_32',
  bldKeep: 'PX2D_icon_bld_keep_32',
  bldLumber: 'PX2D_icon_bld_lumber_32',
  bldWarehouse: 'PX2D_icon_bld_warehouse_32',
  formation: 'PX2D_icon_formation_32',
  resFood: 'PX2D_icon_res_food_32',
  resGold: 'PX2D_icon_res_gold_32',
  resIron: 'PX2D_icon_res_iron_32',
  resStone: 'PX2D_icon_res_stone_32',
  resWood: 'PX2D_icon_res_wood_32',
} as const;

export type IconAssetId = (typeof ICON_ASSET_IDS)[keyof typeof ICON_ASSET_IDS];

const ALL_ASSET_IDS: IconAssetId[] = Object.values(ICON_ASSET_IDS);

function resourcePath(assetId: string): string {
  return `textures/2d/icons/${assetId}`;
}

/**
 * Lazy-loads PX2D icons from resources/textures/2d/icons.
 * Missing assets resolve to null; callers should hide Sprite or keep placeholder — no throw.
 */
export class IconSpriteRegistry {
  private static frames = new Map<string, SpriteFrame | null>();
  private static loadPromise: Promise<void> | null = null;

  static preload(): Promise<void> {
    if (this.loadPromise) {
      return this.loadPromise;
    }
    this.loadPromise = Promise.all(
      ALL_ASSET_IDS.map(
        (assetId) =>
          new Promise<void>((resolve) => {
            const path = resourcePath(assetId);
            resources.load(path, SpriteFrame, (err, sf) => {
              if (err || !sf) {
                console.warn(`[IconSpriteRegistry] icon missing: ${path}`, err?.message ?? '');
                this.frames.set(assetId, null);
              } else {
                applyPixelArtSampling(sf);
                this.frames.set(assetId, sf);
              }
              resolve();
            });
          }),
      ),
    ).then(() => undefined);
    return this.loadPromise;
  }

  /** @param assetId 檔名無擴展名，例 `PX2D_icon_res_food_32` 或 `ICON_ASSET_IDS.resFood` */
  static getSpriteFrame(assetId: string): SpriteFrame | null {
    const cached = this.frames.get(assetId);
    if (cached !== undefined) {
      return cached;
    }
    return null;
  }

  /**
   * 將圖標套到 Sprite；缺圖時關閉 Sprite（不 crash）。
   * @returns 是否成功套用貼圖
   */
  static applyIconToSprite(sprite: Sprite | null, assetId: string, displaySize = 32): boolean {
    if (!sprite) {
      return false;
    }
    const sf = this.getSpriteFrame(assetId);
    if (!sf) {
      sprite.enabled = false;
      sprite.spriteFrame = null;
      return false;
    }
    sprite.enabled = true;
    sprite.sizeMode = Sprite.SizeMode.CUSTOM;
    sprite.spriteFrame = sf;
    const ui = sprite.node.getComponent(UITransform) ?? sprite.node.addComponent(UITransform);
    ui.setContentSize(displaySize, displaySize);
    return true;
  }

  static hasAnyIcon(): boolean {
    for (const sf of this.frames.values()) {
      if (sf) {
        return true;
      }
    }
    return false;
  }

  static isSpriteLoaded(assetId: string): boolean {
    const sf = this.frames.get(assetId);
    return sf != null;
  }
}
