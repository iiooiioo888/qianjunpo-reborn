import { _decorator, Color, Component, Graphics, Node, Sprite, UITransform } from 'cc';
import { ViewUnit } from '../logic/TacticalSnapshot';
import { drawUnitSelectionRing } from './BoardSelectionVisuals';
import { ownerAccentColor } from './TerrainPalette';
import { boardUnitDisplaySize, UNIT_ART_CANVAS_PX } from './PixelSpriteUtil';
import { UNIT_TEXTURE_KEYS, UnitSpriteRegistry } from './UnitSpriteRegistry';

const { ccclass, property } = _decorator;

/**
 * 戰術單位顯示：優先 `UnitSpriteRegistry`（鍵 infantry/cavalry 等）+ Nearest；
 * 缺圖時退回幾何占位。v03 128×128 畫布時仍用整數倍縮入 cellSize。
 */
@ccclass('UnitPlaceholderView')
export class UnitPlaceholderView extends Component {
  @property
  cellSize = 32;

  private g: Graphics | null = null;
  private ringGfx: Graphics | null = null;
  private sprite: Sprite | null = null;
  private selected = false;

  onLoad(): void {
    const ui = this.getComponent(UITransform) ?? this.addComponent(UITransform);
    ui.setContentSize(this.cellSize, this.cellSize);
    this.g = this.getComponent(Graphics) ?? this.addComponent(Graphics);
    const ringNode = new Node('SelectionRing');
    ringNode.setParent(this.node);
    this.ringGfx = ringNode.addComponent(Graphics);
  }

  setSelected(on: boolean): void {
    this.selected = on;
    this.redrawSelectionRing();
  }

  drawUnit(unit: ViewUnit): void {
    if (!this.sprite) {
      this.sprite = this.getComponent(Sprite) ?? this.addComponent(Sprite);
      this.sprite.sizeMode = Sprite.SizeMode.CUSTOM;
    }
    const applied = UnitSpriteRegistry.applyUnitToSprite(
      this.sprite,
      unit.type,
      this.cellSize,
      UNIT_ART_CANVAS_PX,
    );
    if (applied) {
      if (this.g) {
        this.g.clear();
        this.g.enabled = false;
      }
      this.redrawSelectionRing();
      return;
    }
    this.drawGraphicsPlaceholder(unit);
  }

  private redrawSelectionRing(): void {
    const g = this.ringGfx;
    if (!g) {
      return;
    }
    g.clear();
    if (!this.selected) {
      return;
    }
    drawUnitSelectionRing(g, this.cellSize);
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
    const artDisplay = boardUnitDisplaySize(UNIT_ART_CANVAS_PX, s);
    const pad = Math.max(2, Math.floor((s - artDisplay) / 2));
    const fill = ownerAccentColor(unit.owner);
    g.fillColor = fill;
    g.rect(pad, pad, s - pad * 2, s - pad * 2);
    g.fill();
    g.lineWidth = 1;
    g.strokeColor = new Color(20, 20, 28, 255);
    g.rect(pad, pad, s - pad * 2, s - pad * 2);
    g.stroke();
    g.fillColor = new Color(255, 255, 255, 180);
    const key = UnitSpriteRegistry.resolveKeyForUnitType(unit.type);
    if (key === UNIT_TEXTURE_KEYS.cavalry) {
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
    this.redrawSelectionRing();
  }
}
