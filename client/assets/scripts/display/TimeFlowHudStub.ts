import { _decorator, Component, Label, Node, UITransform } from 'cc';
import { timeFlowRateToFloat, ViewSnapshot } from '../logic/TacticalSnapshot';

const { ccclass, property } = _decorator;

/** Stub HUD hook for pkg/timedilation time_flow_rate (display only). */
@ccclass('TimeFlowHudStub')
export class TimeFlowHudStub extends Component {
  @property(Label)
  rateLabel: Label | null = null;

  @property(Label)
  frameLabel: Label | null = null;

  onLoad(): void {
    if (!this.rateLabel) {
      const n = new Node('RateLabel');
      n.setParent(this.node);
      n.setPosition(0, 0, 0);
      const ui = n.addComponent(UITransform);
      ui.setContentSize(400, 40);
      this.rateLabel = n.addComponent(Label);
      this.rateLabel.fontSize = 20;
      this.rateLabel.lineHeight = 24;
    }
    if (!this.frameLabel) {
      const n = new Node('FrameLabel');
      n.setParent(this.node);
      n.setPosition(0, -28, 0);
      const ui = n.addComponent(UITransform);
      ui.setContentSize(400, 40);
      this.frameLabel = n.addComponent(Label);
      this.frameLabel.fontSize = 18;
      this.frameLabel.lineHeight = 22;
    }
  }

  bindSnapshot(snap: ViewSnapshot): void {
    const rate = timeFlowRateToFloat(snap.timeFlowRateParts);
    if (this.rateLabel) {
      this.rateLabel.string = `time_flow_rate: ${rate.toFixed(2)}x (${snap.timeFlowRateParts}/10000)`;
    }
    if (this.frameLabel) {
      this.frameLabel.string = `lockstep frame: ${snap.lockstepFrame}  hash: ${snap.initialStateHash}`;
    }
  }

  /** Future: stream per-frame rates from replay v2 TimeFlowRates or Janus heartbeat. */
  setTimeFlowRateParts(parts: number): void {
    if (this.rateLabel) {
      this.rateLabel.string = `time_flow_rate: ${timeFlowRateToFloat(parts).toFixed(2)}x (${parts}/10000)`;
    }
  }
}
