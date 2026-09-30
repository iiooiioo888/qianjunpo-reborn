import { _decorator, Color, Component, Graphics, UITransform } from 'cc';
import { ViewUnit } from '../logic/TacticalSnapshot';
import { ownerAccentColor } from './TerrainPalette';

const { ccclass, property } = _decorator;

/** 1px-outline cartoon placeholder until PX2D unit sprites import from art/2d. */
@ccclass('UnitPlaceholderView')
export class UnitPlaceholderView extends Component {
  @property
  cellSize = 32;

  private g: Graphics | null = null;

  onLoad(): void {
    const ui = this.getComponent(UITransform) ?? this.addComponent(UITransform);
    ui.setContentSize(this.cellSize, this.cellSize);
    this.g = this.getComponent(Graphics) ?? this.addComponent(Graphics);
  }

  drawUnit(unit: ViewUnit): void {
    const g = this.g;
    if (!g) {
      return;
    }
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
    // Type notch (infantry square / cavalry diamond hint)
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
