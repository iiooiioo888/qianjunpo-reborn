import { _decorator, Component, JsonAsset, Node, resources, Widget } from 'cc';
import { TacticalBoardView } from '../display/TacticalBoardView';
import { TimeFlowHudStub } from '../display/TimeFlowHudStub';

const { ccclass, property } = _decorator;

@ccclass('TacticalBootstrap')
export class TacticalBootstrap extends Component {
  /** resources/ relative path without extension */
  @property
  snapshotResource = 'data/tactical/demo_initial';

  onLoad(): void {
    const boardNode = new Node('Board');
    boardNode.setParent(this.node);
    const boardView = boardNode.addComponent(TacticalBoardView);
    boardView.cellSize = 32;

    const hudNode = new Node('TimeFlowHud');
    hudNode.setParent(this.node);
    hudNode.setPosition(-320, 300, 0);
    const hud = hudNode.addComponent(TimeFlowHudStub);
    const hudWidget = hudNode.addComponent(Widget);
    hudWidget.isAlignTop = true;
    hudWidget.isAlignLeft = true;
    hudWidget.top = 16;
    hudWidget.left = 16;

    resources.load(this.snapshotResource, JsonAsset, (err, asset) => {
      if (err || !asset) {
        console.error('[TacticalBootstrap] failed to load snapshot', err);
        return;
      }
      boardView.applySnapshot(asset.json);
      const snap = boardView.getSnapshot();
      if (snap) {
        hud.bindSnapshot(snap);
      }
    });
  }
}
