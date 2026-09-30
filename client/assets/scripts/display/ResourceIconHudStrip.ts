import { _decorator, Component, Node, Sprite, UITransform } from 'cc';
import { ICON_ASSET_IDS, IconSpriteRegistry } from './IconSpriteRegistry';

const { ccclass, property } = _decorator;

/** 開發用：在 HUD 旁展示五種資源 PX2D 圖標（32px 整數倍，Nearest）。 */
@ccclass('ResourceIconHudStrip')
export class ResourceIconHudStrip extends Component {
  @property
  iconSize = 32;

  @property
  gap = 4;

  private built = false;

  buildStrip(): void {
    if (this.built) {
      return;
    }
    this.built = true;
    const ids = [
      ICON_ASSET_IDS.resFood,
      ICON_ASSET_IDS.resWood,
      ICON_ASSET_IDS.resStone,
      ICON_ASSET_IDS.resIron,
      ICON_ASSET_IDS.resGold,
    ];
    let x = 0;
    for (const assetId of ids) {
      const n = new Node(assetId);
      n.setParent(this.node);
      n.setPosition(x, 0, 0);
      const ui = n.addComponent(UITransform);
      ui.setContentSize(this.iconSize, this.iconSize);
      const sp = n.addComponent(Sprite);
      IconSpriteRegistry.applyIconToSprite(sp, assetId, this.iconSize);
      x += this.iconSize + this.gap;
    }
  }
}
