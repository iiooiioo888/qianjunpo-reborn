import { _decorator, Color, Component, Graphics, Label, Node, UITransform } from 'cc';
import {
  formatLockstepFrameLine,
  formatSelectionLine,
  HudLockstepSyncContext,
  mockSyncContextFromBootstrap,
} from './LockstepHudFormat';
import {
  formatSimTimeFromLockstepFrame,
  formatTimeFlowRatePercent,
  isTimeFlowOverload,
  timeFlowBarFillRatio,
  timeFlowRateToFloat,
  ViewSnapshot,
} from '../logic/TacticalSnapshot';

const { ccclass, property } = _decorator;

const RATE_LABEL_NORMAL = new Color(240, 240, 245, 255);
const RATE_LABEL_OVERLOAD = new Color(255, 150, 90, 255);
const RATE_BAR_BG = new Color(40, 44, 58, 255);
const RATE_BAR_FILL_NORMAL = new Color(90, 210, 130, 255);
const RATE_BAR_FILL_OVERLOAD = new Color(255, 120, 70, 255);
const RATE_BAR_WIDTH = 240;
const RATE_BAR_HEIGHT = 8;

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
  selectionLabel: Label | null = null;

  @property(Label)
  statusLabel: Label | null = null;

  @property(Label)
  simTimeLabel: Label | null = null;

  private rateBarFillGfx: Graphics | null = null;

  private lastSelectedUnitId: number | null = null;

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
      this.rateLabel.color = RATE_LABEL_NORMAL;
    }
    this.ensureRateBar();
    if (!this.simTimeLabel) {
      const n = new Node('SimTimeLabel');
      n.setParent(this.node);
      n.setPosition(0, -40, 0);
      const ui = n.addComponent(UITransform);
      ui.setContentSize(480, 24);
      this.simTimeLabel = n.addComponent(Label);
      this.simTimeLabel.fontSize = 17;
      this.simTimeLabel.lineHeight = 21;
    }
    if (!this.frameLabel) {
      const n = new Node('FrameLabel');
      n.setParent(this.node);
      n.setPosition(0, -58, 0);
      const ui = n.addComponent(UITransform);
      ui.setContentSize(480, 40);
      this.frameLabel = n.addComponent(Label);
      this.frameLabel.fontSize = 18;
      this.frameLabel.lineHeight = 22;
    }
    if (!this.selectionLabel) {
      const n = new Node('SelectionLabel');
      n.setParent(this.node);
      n.setPosition(0, -86, 0);
      const ui = n.addComponent(UITransform);
      ui.setContentSize(480, 40);
      this.selectionLabel = n.addComponent(Label);
      this.selectionLabel.fontSize = 18;
      this.selectionLabel.lineHeight = 22;
    }
    if (!this.statusLabel) {
      const n = new Node('StatusLabel');
      n.setParent(this.node);
      n.setPosition(0, -114, 0);
      const ui = n.addComponent(UITransform);
      ui.setContentSize(520, 36);
      this.statusLabel = n.addComponent(Label);
      this.statusLabel.fontSize = 16;
      this.statusLabel.lineHeight = 20;
      this.statusLabel.color = new Color(255, 180, 100, 255);
    }
    const retryN = new Node('LivePrepareRetry');
    retryN.setParent(this.node);
    retryN.setPosition(0, -146, 0);
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

  private ensureRateBar(): void {
    if (this.rateBarFillGfx) {
      return;
    }
    const barRoot = new Node('RateBar');
    barRoot.setParent(this.node);
    barRoot.setPosition(0, -26, 0);
    const rootUi = barRoot.addComponent(UITransform);
    rootUi.setContentSize(RATE_BAR_WIDTH, RATE_BAR_HEIGHT);
    const bg = barRoot.addComponent(Graphics);
    bg.fillColor = RATE_BAR_BG;
    bg.rect(0, 0, RATE_BAR_WIDTH, RATE_BAR_HEIGHT);
    bg.fill();
    const fillNode = new Node('RateBarFill');
    fillNode.setParent(barRoot);
    fillNode.setPosition(0, 0, 0);
    fillNode.addComponent(UITransform).setContentSize(RATE_BAR_WIDTH, RATE_BAR_HEIGHT);
    this.rateBarFillGfx = fillNode.addComponent(Graphics);
  }

  private syncTimeFlowVisuals(parts: number, lockstepFrame: number): void {
    const overloaded = isTimeFlowOverload(parts);
    if (this.rateLabel) {
      const rate = timeFlowRateToFloat(parts);
      const pct = formatTimeFlowRatePercent(parts);
      this.rateLabel.string = `time_flow_rate: ${rate.toFixed(2)}x (${parts}/10000) · ${pct}`;
      this.rateLabel.color = overloaded ? RATE_LABEL_OVERLOAD : RATE_LABEL_NORMAL;
    }
    if (this.simTimeLabel) {
      this.simTimeLabel.string = formatSimTimeFromLockstepFrame(lockstepFrame);
    }
    if (this.rateBarFillGfx) {
      const fillW = RATE_BAR_WIDTH * timeFlowBarFillRatio(parts);
      const g = this.rateBarFillGfx;
      g.clear();
      g.fillColor = overloaded ? RATE_BAR_FILL_OVERLOAD : RATE_BAR_FILL_NORMAL;
      if (fillW > 0) {
        g.rect(0, 0, fillW, RATE_BAR_HEIGHT);
        g.fill();
      }
    }
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
    this.syncSelectionLabel();
  }

  /** Local board selection (grid x,y from latest snapshot). */
  setSelectedUnitId(unitId: number | null): void {
    this.lastSelectedUnitId = unitId;
    this.syncSelectionLabel();
  }

  private syncSelectionLabel(): void {
    if (!this.selectionLabel) {
      return;
    }
    this.selectionLabel.string = formatSelectionLine(this.lastSnapshot, this.lastSelectedUnitId);
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
    this.syncTimeFlowVisuals(snap.timeFlowRateParts, snap.lockstepFrame);
    if (this.frameLabel) {
      this.frameLabel.string = formatLockstepFrameLine(snap, this.lastSync);
    }
    this.syncSelectionLabel();
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
