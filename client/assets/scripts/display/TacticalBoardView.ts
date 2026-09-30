import { _decorator, Color, Component, Graphics, Node, UITransform } from 'cc';
import { shouldRedrawGrid, shouldRedrawUnits } from '../logic/SnapshotDisplayDiff';
import { parseViewSnapshot, ViewSnapshot } from '../logic/TacticalSnapshot';
import { terrainFillColor } from './TerrainPalette';
import { UnitPlaceholderView } from './UnitPlaceholderView';

const { ccclass, property } = _decorator;

@ccclass('TacticalBoardView')
export class TacticalBoardView extends Component {
  @property
  cellSize = 32;

  private gridGfx: Graphics | null = null;
  private unitLayer: Node | null = null;
  private snapshot: ViewSnapshot | null = null;

  onLoad(): void {
    const ui = this.getComponent(UITransform) ?? this.addComponent(UITransform);
    const n = 19;
    ui.setContentSize(n * this.cellSize, n * this.cellSize);
    this.gridGfx = this.getComponent(Graphics) ?? this.addComponent(Graphics);
    this.unitLayer = new Node('Units');
    this.unitLayer.setParent(this.node);
  }

  applySnapshot(raw: unknown): void {
    const next = parseViewSnapshot(raw);
    const prev = this.snapshot;
    this.snapshot = next;
    if (shouldRedrawGrid(prev, next)) {
      this.redrawGrid();
    }
    if (shouldRedrawUnits(prev, next)) {
      this.redrawUnits();
    }
  }

  getSnapshot(): ViewSnapshot | null {
    return this.snapshot;
  }

  private redrawGrid(): void {
    const snap = this.snapshot;
    const g = this.gridGfx;
    if (!snap || !g) {
      return;
    }
    g.clear();
    const cs = this.cellSize;
    for (let y = 0; y < snap.boardSize; y++) {
      for (let x = 0; x < snap.boardSize; x++) {
        const cell = snap.cells[y][x];
        g.fillColor = terrainFillColor(cell.terrain);
        const px = x * cs;
        const py = (snap.boardSize - 1 - y) * cs;
        g.rect(px, py, cs - 1, cs - 1);
        g.fill();
        if (!cell.passable) {
          g.strokeColor = new Color(40, 40, 48, 200);
          g.lineWidth = 1;
          g.rect(px + 0.5, py + 0.5, cs - 2, cs - 2);
          g.stroke();
        }
      }
    }
    g.strokeColor = new Color(30, 30, 36, 255);
    g.lineWidth = 1;
    const edge = snap.boardSize * cs;
    g.rect(0, 0, edge, edge);
    g.stroke();
  }

  private redrawUnits(): void {
    const snap = this.snapshot;
    const layer = this.unitLayer;
    if (!snap || !layer) {
      return;
    }
    layer.removeAllChildren();
    const cs = this.cellSize;
    for (const unit of snap.units) {
      if (unit.hp <= 0) {
        continue;
      }
      const node = new Node(`unit_${unit.id}`);
      node.setParent(layer);
      const px = unit.x * cs + cs / 2;
      const py = (snap.boardSize - 1 - unit.y) * cs + cs / 2;
      node.setPosition(px, py, 0);
      const view = node.addComponent(UnitPlaceholderView);
      view.cellSize = cs;
      view.drawUnit(unit);
    }
  }
}
