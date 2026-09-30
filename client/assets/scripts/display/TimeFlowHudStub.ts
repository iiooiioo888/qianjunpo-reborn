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
    retryLbl.string = '';
    retryLbl.fontSize = 16;
    retryLbl.lineHeight = 20;
    retryLbl.color = new Color(120, 200, 255, 255);
    retryN.active = false;
    retryN.on(Node.EventType.TOUCH_END, this.onLiveRetryTap, this);
    this.retryTapNode = retryN;
    this.retryLabel = retryLbl;
  }

  private onLiveRetryTap(): void {
    const fn = this.liveCommandRetry ?? this.livePollReconnect ?? this.livePrepareRetry;
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
  private livePollReconnect: (() => void) | null = null;
  private liveCommandRetry: (() => void) | null = null;
  private retryTapNode: Node | null = null;
  private retryLabel: Label | null = null;

  private static readonly LABEL_PREPARE_RETRY = '重試 Live 建局（connect → enter-battle）';
  private static readonly LABEL_POLL_RECONNECT = '重連 Live';
  private static readonly LABEL_COMMAND_RETRY = '重試戰術指令';

  private syncLiveRetryAffordance(): void {
    const command = this.liveCommandRetry;
    const reconnect = this.livePollReconnect;
    const prepare = this.livePrepareRetry;
    const show = command != null || reconnect != null || prepare != null;
    if (this.retryTapNode) {
      this.retryTapNode.active = show;
    }
    if (this.retryLabel && show) {
      this.retryLabel.string = command
        ? TimeFlowHudStub.LABEL_COMMAND_RETRY
        : reconnect
          ? TimeFlowHudStub.LABEL_POLL_RECONNECT
          : TimeFlowHudStub.LABEL_PREPARE_RETRY;
    }
  }

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
   * `livePrepareRetry`：connect→enter-battle 建局失敗時一鍵重試。
   * `livePollReconnect`：Live 快照輪詢失敗時一鍵重連（停 poller → 重新建局 → 恢復輪詢）。
   * `liveCommandRetry`：戰術指令拒絕／HTTP／網路失敗時重送同一筆指令（不假樂觀改盤）。
   */
  setNetworkStatus(
    message: string | null,
    opts?: {
      livePrepareRetry?: () => void;
      livePollReconnect?: () => void;
      liveCommandRetry?: () => void;
      /** 僅更新第三行文字，保留既有重試按鈕（例如快照輪詢成功但指令仍失敗）。 */
      preserveRetryAffordance?: boolean;
    },
  ): void {
    if (this.statusLabel) {
      this.statusLabel.string = message ?? '';
    }
    if (opts?.preserveRetryAffordance) {
      this.syncLiveRetryAffordance();
      return;
    }
    if (opts?.liveCommandRetry) {
      this.liveCommandRetry = opts.liveCommandRetry;
      this.livePollReconnect = null;
      this.livePrepareRetry = null;
    } else if (opts?.livePollReconnect) {
      this.livePollReconnect = opts.livePollReconnect;
      this.liveCommandRetry = null;
      this.livePrepareRetry = null;
    } else if (opts?.livePrepareRetry) {
      this.livePrepareRetry = opts.livePrepareRetry;
      this.liveCommandRetry = null;
      this.livePollReconnect = null;
    } else if (opts === undefined) {
      this.livePrepareRetry = null;
      this.livePollReconnect = null;
      this.liveCommandRetry = null;
    }
    this.syncLiveRetryAffordance();
  }

  clearLiveCommandRetry(): void {
    this.liveCommandRetry = null;
    this.syncLiveRetryAffordance();
  }

  hasLiveCommandRetry(): boolean {
    return this.liveCommandRetry != null;
  }
}
