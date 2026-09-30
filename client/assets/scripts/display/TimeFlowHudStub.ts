import { _decorator, Color, Component, Label, Node, UITransform } from 'cc';
import {
  formatLockstepFrameLine,
  HudLockstepSyncContext,
  mockSyncContextFromBootstrap,
} from './LockstepHudFormat';
import { timeFlowRateToFloat, ViewSnapshot } from '../logic/TacticalSnapshot';

const { ccclass, property } = _decorator;

/**
 * HUD：速率與幀號僅來自 ViewSnapshot（Roma live `timeFlowRateParts`，萬分比 10000=1.0x）。
 * 不自行插值或造假 time_flow_rate。
 */
@ccclass('TimeFlowHudStub')
export class TimeFlowHudStub extends Component {
  private lastSync: HudLockstepSyncContext = mockSyncContextFromBootstrap('connected');

  @property(Label)
  rateLabel: Label | null = null;

  @property(Label)
  frameLabel: Label | null = null;

  @property(Label)
  statusLabel: Label | null = null;

  onLoad(): void {
    if (!this.rateLabel) {
      const n = new Node('RateLabel');
      n.setParent(this.node);
      n.setPosition(0, 0, 0);
      const ui = n.addComponent(UITransform);
      ui.setContentSize(480, 40);
      this.rateLabel = n.addComponent(Label);
      this.rateLabel.fontSize = 20;
      this.rateLabel.lineHeight = 24;
    }
    if (!this.frameLabel) {
      const n = new Node('FrameLabel');
      n.setParent(this.node);
      n.setPosition(0, -28, 0);
      const ui = n.addComponent(UITransform);
      ui.setContentSize(480, 40);
      this.frameLabel = n.addComponent(Label);
      this.frameLabel.fontSize = 18;
      this.frameLabel.lineHeight = 22;
    }
    if (!this.statusLabel) {
      const n = new Node('StatusLabel');
      n.setParent(this.node);
      n.setPosition(0, -56, 0);
      const ui = n.addComponent(UITransform);
      ui.setContentSize(520, 36);
      this.statusLabel = n.addComponent(Label);
      this.statusLabel.fontSize = 16;
      this.statusLabel.lineHeight = 20;
      this.statusLabel.color = new Color(255, 180, 100, 255);
    }
    const retryN = new Node('LivePrepareRetry');
    retryN.setParent(this.node);
    retryN.setPosition(0, -88, 0);
    const retryUi = retryN.addComponent(UITransform);
    retryUi.setContentSize(520, 28);
    const retryLbl = retryN.addComponent(Label);
    retryLbl.string = '重試 Live 建局（connect → enter-battle）';
    retryLbl.fontSize = 16;
    retryLbl.lineHeight = 20;
    retryLbl.color = new Color(120, 200, 255, 255);
    retryN.active = false;
    retryN.on(Node.EventType.TOUCH_END, this.onLivePrepareRetryTap, this);
    this.retryTapNode = retryN;
  }

  private onLivePrepareRetryTap(): void {
    const fn = this.livePrepareRetry;
    if (!fn) {
      return;
    }
    fn();
  }

  /** 更新 mock／live 同步語境（影響 lockstep 第二行，不覆寫第三行 overlay）。 */
  setLockstepSyncContext(ctx: HudLockstepSyncContext): void {
    this.lastSync = ctx;
    if (this.frameLabel && this.lastSnapshot) {
      this.frameLabel.string = formatLockstepFrameLine(this.lastSnapshot, this.lastSync);
    }
  }

  private lastSnapshot: ViewSnapshot | null = null;
  private livePrepareRetry: (() => void) | null = null;
  private retryTapNode: Node | null = null;

  /** 每次 mock／live 快照成功後呼叫；`timeFlowRateParts` 為 Roma 權威顯示值。 */
  updateFromSnapshot(snap: ViewSnapshot, sync?: HudLockstepSyncContext): void {
    this.lastSnapshot = snap;
    if (sync) {
      this.lastSync = sync;
    }
    const rate = timeFlowRateToFloat(snap.timeFlowRateParts);
    if (this.rateLabel) {
      this.rateLabel.string = `time_flow_rate: ${rate.toFixed(2)}x (${snap.timeFlowRateParts}/10000)`;
    }
    if (this.frameLabel) {
      this.frameLabel.string = formatLockstepFrameLine(snap, this.lastSync);
    }
  }

  /** @deprecated 使用 updateFromSnapshot */
  bindSnapshot(snap: ViewSnapshot): void {
    this.updateFromSnapshot(snap);
  }

  /**
   * Live 輪詢／建局狀態 overlay（第三行）。
   * `livePrepareRetry`：僅在 connect→enter-battle 建局失敗時顯示一鍵重試（不造假棋盤狀態）。
   */
  setNetworkStatus(
    message: string | null,
    opts?: { livePrepareRetry?: () => void },
  ): void {
    if (this.statusLabel) {
      this.statusLabel.string = message ?? '';
    }
    if (opts?.livePrepareRetry) {
      this.livePrepareRetry = opts.livePrepareRetry;
      if (this.retryTapNode) {
        this.retryTapNode.active = true;
      }
    } else {
      this.livePrepareRetry = null;
      if (this.retryTapNode) {
        this.retryTapNode.active = false;
      }
    }
  }
}
