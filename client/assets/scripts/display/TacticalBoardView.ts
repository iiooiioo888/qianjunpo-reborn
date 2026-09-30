import { _decorator, Color, Component, Graphics, Node, UITransform } from 'cc';
import { Coord } from '../logic/BoardCoord';
import { shouldRedrawGrid, shouldRedrawUnits } from '../logic/SnapshotDisplayDiff';
import { BOARD_SIZE, parseViewSnapshot, ViewSnapshot } from '../logic/TacticalSnapshot';
import { terrainFillColor } from './TerrainPalette';
import { UnitPlaceholderView } from './UnitPlaceholderView';

const { ccclass, property } = _decorator;

@ccclass('TacticalBoardView')
export class TacticalBoardView extends Component {
  @property
  cellSize = 32;

  private gridGfx: Graphics | null = null;
  private highlightGfx: Graphics | null = null;
  private unitLayer: Node | null = null;
  private snapshot: ViewSnapshot | null = null;
  private selectedUnitId: number | null = null;
  private legalDestinations: Coord[] = [];

  onLoad(): void {
    const ui = this.getComponent(UITransform) ?? this.addComponent(UITransform);
    const n = BOARD_SIZE;
    ui.setContentSize(n * this.cellSize, n * this.cellSize);
    this.gridGfx = this.getComponent(Graphics) ?? this.addComponent(Graphics);
    const highlightNode = new Node('MoveHighlights');
    highlightNode.setParent(this.node);
    this.highlightGfx = highlightNode.addComponent(Graphics);
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
    this.redrawHighlights();
  }

  getSnapshot(): ViewSnapshot | null {
    return this.snapshot;
  }

  setSelection(unitId: number | null, legal: Coord[]): void {
    this.selectedUnitId = unitId;
    this.legalDestinations = legal;
    this.redrawHighlights();
    this.updateUnitSelectionVisuals();
  }

  /** Local board space (origin bottom-left of grid). */
  pixelToGrid(px: number, py: number): Coord | null {
    const snap = this.snapshot;
    if (!snap) {
      return null;
    }
    const cs = this.cellSize;
    const edge = snap.boardSize * cs;
    if (px < 0 || py < 0 || px >= edge || py >= edge) {
      return null;
    }
    const x = Math.floor(px / cs);
    const y = snap.boardSize - 1 - Math.floor(py / cs);
    if (x < 0 || y < 0 || x >= snap.boardSize || y >= snap.boardSize) {
      return null;
    }
    return { x, y };
  }

  private redrawHighlights(): void {
    const g = this.highlightGfx;
    const snap = this.snapshot;
    if (!g || !snap) {
      return;
    }
    g.clear();
    const cs = this.cellSize;
    const inset = 3;
    for (const c of this.legalDestinations) {
      const px = c.x * cs;
      const py = (snap.boardSize - 1 - c.y) * cs;
      g.fillColor = new Color(80, 200, 120, 110);
      g.rect(px + inset, py + inset, cs - inset * 2, cs - inset * 2);
      g.fill();
    }
    if (this.selectedUnitId != null) {
      const unit = snap.units.find((u) => u.id === this.selectedUnitId && u.hp > 0);
      if (unit) {
        const px = unit.x * cs;
        const py = (snap.boardSize - 1 - unit.y) * cs;
        g.lineWidth = 2;
        g.strokeColor = new Color(255, 220, 60, 255);
        g.rect(px + 1, py + 1, cs - 2, cs - 2);
        g.stroke();
      }
    }
  }

  private updateUnitSelectionVisuals(): void {
    const layer = this.unitLayer;
    if (!layer) {
      return;
    }
    for (const child of layer.children) {
      const view = child.getComponent(UnitPlaceholderView);
      if (!view) {
        continue;
      }
      const id = this.parseUnitIdFromNodeName(child.name);
      view.setSelected(id !== null && id === this.selectedUnitId);
    }
  }

  private parseUnitIdFromNodeName(name: string): number | null {
    const m = /^unit_(\d+)$/.exec(name);
    if (!m) {
      return null;
    }
    return Number(m[1]);
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
      view.setSelected(unit.id === this.selectedUnitId);
    }
  }
}
