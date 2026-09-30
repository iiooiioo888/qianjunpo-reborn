import { _decorator, Color, Component, Label, Node, UITransform } from 'cc';
import { timeFlowRateToFloat, ViewSnapshot } from '../logic/TacticalSnapshot';

const { ccclass, property } = _decorator;

/**
 * HUD：速率與幀號僅來自 ViewSnapshot（Roma live `timeFlowRateParts`，萬分比 10000=1.0x）。
 * 不自行插值或造假 time_flow_rate。
 */
@ccclass('TimeFlowHudStub')
export class TimeFlowHudStub extends Component {
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
  }

  /** 每次 mock／live 快照成功後呼叫；`timeFlowRateParts` 為 Roma 權威顯示值。 */
  updateFromSnapshot(snap: ViewSnapshot): void {
    const rate = timeFlowRateToFloat(snap.timeFlowRateParts);
    if (this.rateLabel) {
      this.rateLabel.string = `time_flow_rate: ${rate.toFixed(2)}x (${snap.timeFlowRateParts}/10000)`;
    }
    if (this.frameLabel) {
      this.frameLabel.string = `lockstep frame: ${snap.lockstepFrame}  hash: ${snap.initialStateHash}`;
    }
  }

  /** @deprecated 使用 updateFromSnapshot */
  bindSnapshot(snap: ViewSnapshot): void {
    this.updateFromSnapshot(snap);
  }

  /** Live 輪詢失敗時顯示；成功後傳 null 清除。 */
  setNetworkStatus(message: string | null): void {
    if (!this.statusLabel) {
      return;
    }
    this.statusLabel.string = message ?? '';
  }
}
