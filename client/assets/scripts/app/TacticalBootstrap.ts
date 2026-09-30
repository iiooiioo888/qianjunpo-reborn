import { _decorator, Component, JsonAsset, Node, resources, Widget } from 'cc';
import { TacticalBoardView } from '../display/TacticalBoardView';
import { TimeFlowHudStub } from '../display/TimeFlowHudStub';
import { UnitSpriteRegistry } from '../display/UnitSpriteRegistry';
import { DEFAULT_NETWORK_STUB } from '../network/JanusGatewayStub';
import { LiveViewSnapshotPoller } from '../network/LiveViewSnapshotPoller';

const { ccclass, property } = _decorator;

@ccclass('TacticalBootstrap')
export class TacticalBootstrap extends Component {
  /** When true, load view JSON from Janus HTTP dev mirror instead of resources mock. */
  @property
  useLiveJanus = false;

  /** Roma battle id (e.g. default/0) from EnterBattle — required when useLiveJanus is true. */
  @property
  liveBattleId = 'default/0';

  /**
   * Live 模式 HTTP 輪詢間隔（ms）。預設 333ms（約 3Hz）。
   * 建議 200–500ms，對齊 Janus `GET /v1/tactical/snapshot` 開發鏡像。
   */
  @property
  livePollIntervalMs = 333;

  /** resources/ relative path without extension (mock mode). */
  @property
  snapshotResource = 'data/tactical/demo_initial';

  private poller: LiveViewSnapshotPoller | null = null;

  onDestroy(): void {
    this.poller?.stop();
    this.poller = null;
  }

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

    const applySnapshot = (raw: unknown) => {
      try {
        boardView.applySnapshot(raw);
        const snap = boardView.getSnapshot();
        if (snap) {
          hud.updateFromSnapshot(snap);
          hud.setNetworkStatus(null);
        }
      } catch (err) {
        console.error('[TacticalBootstrap] snapshot parse/apply failed', err);
        hud.setNetworkStatus('快照解析失敗（見 console）');
      }
    };

    void UnitSpriteRegistry.preload().then(() => {
      if (this.useLiveJanus) {
        this.poller = new LiveViewSnapshotPoller({
          cfg: DEFAULT_NETWORK_STUB,
          battleId: this.liveBattleId,
          intervalMs: this.livePollIntervalMs,
          onSnapshot: applySnapshot,
          onError: (err, retryMs) => {
            console.warn('[TacticalBootstrap] live Janus poll failed', err.message, `(retry ~${retryMs}ms)`);
            hud.setNetworkStatus(`Janus: ${err.message}（約 ${retryMs}ms 後重試）`);
          },
        });
        this.poller.start(true);
        return;
      }

      resources.load(this.snapshotResource, JsonAsset, (err, asset) => {
        if (err || !asset) {
          console.error('[TacticalBootstrap] failed to load mock snapshot', err);
          hud.setNetworkStatus('Mock JSON 載入失敗');
          return;
        }
        applySnapshot(asset.json);
      });
    });
  }
}
