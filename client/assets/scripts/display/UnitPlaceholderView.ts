import { _decorator, Color, Component, Graphics, Sprite, UITransform } from 'cc';
import { ViewUnit } from '../logic/TacticalSnapshot';
import { ownerAccentColor } from './TerrainPalette';
import { UnitSpriteRegistry } from './UnitSpriteRegistry';

const { ccclass, property } = _decorator;

/**
 * 戰術單位顯示：優先 `resources/textures/2d/units` PX2D v02 Sprite（Nearest）；
 * 缺圖時退回幾何占位。
 */
@ccclass('UnitPlaceholderView')
export class UnitPlaceholderView extends Component {
  @property
  cellSize = 32;

  private g: Graphics | null = null;
  private sprite: Sprite | null = null;

  onLoad(): void {
    const ui = this.getComponent(UITransform) ?? this.addComponent(UITransform);
    ui.setContentSize(this.cellSize, this.cellSize);
    this.g = this.getComponent(Graphics) ?? this.addComponent(Graphics);
  }

  drawUnit(unit: ViewUnit): void {
    const sf = UnitSpriteRegistry.getSpriteFrame(unit.type);
    if (sf) {
      this.drawSpriteToken(sf);
      return;
    }
    this.drawGraphicsPlaceholder(unit);
  }

  private drawSpriteToken(sf: NonNullable<ReturnType<typeof UnitSpriteRegistry.getSpriteFrame>>): void {
    if (!this.sprite) {
      this.sprite = this.getComponent(Sprite) ?? this.addComponent(Sprite);
      this.sprite.sizeMode = Sprite.SizeMode.CUSTOM;
    }
    this.sprite.enabled = true;
    this.sprite.spriteFrame = sf;
    const ui = this.getComponent(UITransform);
    if (ui) {
      ui.setContentSize(this.cellSize, this.cellSize);
    }
    if (this.g) {
      this.g.clear();
      this.g.enabled = false;
    }
  }

  private drawGraphicsPlaceholder(unit: ViewUnit): void {
    if (this.sprite) {
      this.sprite.enabled = false;
    }
    const g = this.g;
    if (!g) {
      return;
    }
    g.enabled = true;
    g.clear();
    const s = this.cellSize;
    const pad = 4;
    const fill = ownerAccentColor(unit.owner);
    g.fillColor = fill;
    g.rect(pad, pad, s - pad * 2, s - pad * 2);
    g.fill();
    g.lineWidth = 1;
    g.strokeColor = new Color(20, 20, 28, 255);
    g.rect(pad, pad, s - pad * 2, s - pad * 2);
    g.stroke();
    g.fillColor = new Color(255, 255, 255, 180);
    if (unit.type === 2) {
      g.moveTo(s / 2, pad + 2);
      g.lineTo(s - pad - 2, s / 2);
      g.lineTo(s / 2, s - pad - 2);
      g.lineTo(pad + 2, s / 2);
      g.close();
      g.fill();
    } else {
      g.rect(s / 2 - 3, s / 2 - 3, 6, 6);
      g.fill();
    }
  }
}
