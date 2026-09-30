import { _decorator, Component, JsonAsset, Node, resources, Widget } from 'cc';
import { TacticalBoardView } from '../display/TacticalBoardView';
import { TimeFlowHudStub } from '../display/TimeFlowHudStub';
import { DEFAULT_NETWORK_STUB, fetchLiveViewSnapshot } from '../network/JanusGatewayStub';

const { ccclass, property } = _decorator;

@ccclass('TacticalBootstrap')
export class TacticalBootstrap extends Component {
  /** When true, load view JSON from Janus HTTP dev mirror instead of resources mock. */
  @property
  useLiveJanus = false;

  /** Roma battle id (e.g. default/0) from EnterBattle — required when useLiveJanus is true. */
  @property
  liveBattleId = 'default/0';

  /** resources/ relative path without extension (mock mode). */
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

    const apply = (raw: unknown) => {
      boardView.applySnapshot(raw);
      const snap = boardView.getSnapshot();
      if (snap) {
        hud.bindSnapshot(snap);
      }
    };

    if (this.useLiveJanus) {
      fetchLiveViewSnapshot(DEFAULT_NETWORK_STUB, this.liveBattleId)
        .then(apply)
        .catch((err) => console.error('[TacticalBootstrap] live Janus snapshot failed', err));
      return;
    }

    resources.load(this.snapshotResource, JsonAsset, (err, asset) => {
      if (err || !asset) {
        console.error('[TacticalBootstrap] failed to load mock snapshot', err);
        return;
      }
      apply(asset.json);
    });
  }
}
