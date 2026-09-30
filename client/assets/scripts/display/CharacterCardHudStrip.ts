import { _decorator, Color, Component, Graphics, Node, Sprite, UITransform } from 'cc';
import { CHAR_CARD_TEXTURE_KEYS, CharacterCardSpriteRegistry } from './CharacterCardSpriteRegistry';
import { CHAR_CARD_ART_HEIGHT_PX, CHAR_CARD_ART_WIDTH_PX, charCardDisplaySize } from './PixelSpriteUtil';

const { ccclass, property } = _decorator;

/** 與 {@link CharacterCardSpriteRegistry} 預覽鍵一致（mock 聯調固定三格）。 */
export const CHARACTER_CARD_HUD_PREVIEW_KEYS = [
  CHAR_CARD_TEXTURE_KEYS.char_caocao,
  CHAR_CARD_TEXTURE_KEYS.char_zhangfei,
  CHAR_CARD_TEXTURE_KEYS.char_wu_placeholder,
] as const;

let loggedCardPlaceholderHint = false;

/** 開發用：HUD 展示已登記角色卡（320×400 整數倍縮放，Nearest）；缺圖時色塊占位。 */
@ccclass('CharacterCardHudStrip')
export class CharacterCardHudStrip extends Component {
  @property
  previewMaxWidth = 48;

  @property
  previewMaxHeight = 60;

  @property
  gap = 6;

  private built = false;

  buildStrip(): void {
    if (this.built) {
      return;
    }
    this.built = true;
    const keys = [...CHARACTER_CARD_HUD_PREVIEW_KEYS];
    const { width, height } = charCardDisplaySize(this.previewMaxWidth, this.previewMaxHeight);
    let x = 0;
    for (const key of keys) {
      const n = new Node(key);
      n.setParent(this.node);
      n.setPosition(x, 0, 0);
      const ui = n.addComponent(UITransform);
      ui.setContentSize(width, height);
      const sp = n.addComponent(Sprite);
      const ok = CharacterCardSpriteRegistry.applyCardToSprite(
        sp,
        key,
        this.previewMaxWidth,
        this.previewMaxHeight,
      );
      if (!ok) {
        this.drawCardPlaceholder(n, width, height, key);
      }
      x += width + this.gap;
    }
  }

  private drawCardPlaceholder(node: Node, w: number, h: number, key: string): void {
    const g = node.addComponent(Graphics);
    g.fillColor = new Color(48, 52, 72, 255);
    g.rect(0, 0, w, h);
    g.fill();
    g.lineWidth = 1;
    g.strokeColor = new Color(120, 130, 160, 255);
    g.rect(1, 1, w - 2, h - 2);
    g.stroke();
    g.fillColor = new Color(200, 200, 220, 200);
    const labelH = Math.min(8, Math.floor(h * 0.12));
    g.rect(4, h - labelH - 4, w - 8, labelH);
    g.fill();
    if (!loggedCardPlaceholderHint) {
      loggedCardPlaceholderHint = true;
      console.info(
        `[CharacterCardHudStrip] 缺圖占位（含 ${key} 等）— PNG 放 textures/2d/chars/ ${CHAR_CARD_ART_WIDTH_PX}×${CHAR_CARD_ART_HEIGHT_PX} 或 TextureRegistryDev.applyDevTextureStemOverrides`,
      );
    }
  }
}
