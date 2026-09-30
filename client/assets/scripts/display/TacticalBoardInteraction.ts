import { _decorator, Component, EventTouch, Node, UITransform, Vec2 } from 'cc';
import { Coord } from '../logic/BoardCoord';
import { computeLegalMoveDestinations } from '../logic/ClientMoveReachability';
import { TacticalBoardView } from './TacticalBoardView';

const { ccclass, property } = _decorator;

export type MoveSubmitHandler = (unitId: number, to: Coord) => void | Promise<void>;

/**
 * Tap select unit → highlight legal moves → tap cell to submit move.
 */
@ccclass('TacticalBoardInteraction')
export class TacticalBoardInteraction extends Component {
  @property
  localPlayerId = 0;

  private boardView: TacticalBoardView | null = null;
  private onSubmitMove: MoveSubmitHandler | null = null;
  private selectedUnitId: number | null = null;

  bind(boardView: TacticalBoardView, onSubmitMove: MoveSubmitHandler): void {
    this.boardView = boardView;
    this.onSubmitMove = onSubmitMove;
  }

  onLoad(): void {
    this.node.on(Node.EventType.TOUCH_END, this.onTouchEnd, this);
  }

  onDestroy(): void {
    this.node.off(Node.EventType.TOUCH_END, this.onTouchEnd, this);
  }

  clearSelection(): void {
    this.selectedUnitId = null;
    this.boardView?.setSelection(null, []);
  }

  /** Recompute highlights after snapshot refresh. */
  refreshSelectionFromSnapshot(): void {
    if (this.selectedUnitId == null || !this.boardView) {
      return;
    }
    const snap = this.boardView.getSnapshot();
    if (!snap) {
      this.clearSelection();
      return;
    }
    const unit = snap.units.find((u) => u.id === this.selectedUnitId && u.hp > 0);
    if (!unit || unit.owner !== this.localPlayerId) {
      this.clearSelection();
      return;
    }
    const legal = computeLegalMoveDestinations(snap, unit.id);
    this.boardView.setSelection(unit.id, legal);
  }

  private onTouchEnd(ev: EventTouch): void {
    const board = this.boardView;
    if (!board) {
      return;
    }
    const snap = board.getSnapshot();
    if (!snap) {
      return;
    }
    const ui = this.node.getComponent(UITransform);
    if (!ui) {
      return;
    }
    const loc = ev.getUILocation();
    const local = new Vec2();
    ui.convertToNodeSpaceAR(new Vec2(loc.x, loc.y), local);
    const grid = board.pixelToGrid(local.x, local.y);
    if (!grid) {
      return;
    }
    const cell = snap.cells[grid.y][grid.x];
    const unitAt = snap.units.find((u) => u.id === cell.unitId && u.hp > 0);

    if (this.selectedUnitId != null) {
      const legal = computeLegalMoveDestinations(snap, this.selectedUnitId);
      const isLegal = legal.some((c) => c.x === grid.x && c.y === grid.y);
      if (isLegal) {
        void this.onSubmitMove?.(this.selectedUnitId, grid);
        return;
      }
    }

    if (unitAt && unitAt.owner === this.localPlayerId) {
      this.selectedUnitId = unitAt.id;
      const legal = computeLegalMoveDestinations(snap, unitAt.id);
      board.setSelection(unitAt.id, legal);
      return;
    }

    this.clearSelection();
  }
}
