import { fetchLiveViewSnapshot, TacticalNetworkConfig } from './JanusGatewayStub';

export interface LiveSnapshotPollerOptions {
  cfg: TacticalNetworkConfig;
  battleId: string;
  sessionId?: string;
  /** Polling period on success (ms). Default 333 ≈ 3 Hz; use 200–500 for live HUD. */
  intervalMs?: number;
  onSnapshot: (raw: unknown) => void;
  onError?: (err: Error, retryDelayMs: number) => void;
}

/**
 * Schedules Janus HTTP snapshot polls for browser preview (useLiveJanus).
 * Uses setTimeout (not setInterval) so backoff can extend delay after failures.
 */
export class LiveViewSnapshotPoller {
  private readonly intervalMs: number;
  private timeoutId: ReturnType<typeof setTimeout> | null = null;
  private stopped = false;
  private inFlight = false;
  private backoffMs = 0;

  constructor(private readonly opts: LiveSnapshotPollerOptions) {
    this.intervalMs = Math.max(50, opts.intervalMs ?? 333);
  }

  start(immediate = true): void {
    this.stopped = false;
    if (immediate) {
      void this.tick();
    } else {
      this.schedule(this.intervalMs);
    }
  }

  stop(): void {
    this.stopped = true;
    if (this.timeoutId !== null) {
      clearTimeout(this.timeoutId);
      this.timeoutId = null;
    }
  }

  private schedule(delayMs: number): void {
    if (this.stopped) {
      return;
    }
    if (this.timeoutId !== null) {
      clearTimeout(this.timeoutId);
    }
    this.timeoutId = setTimeout(() => void this.tick(), delayMs);
  }

  private async tick(): Promise<void> {
    if (this.stopped || this.inFlight) {
      this.schedule(this.intervalMs);
      return;
    }
    this.inFlight = true;
    try {
      const raw = await fetchLiveViewSnapshot(
        this.opts.cfg,
        this.opts.battleId,
        this.opts.sessionId ?? '',
      );
      this.backoffMs = 0;
      this.opts.onSnapshot(raw);
      this.schedule(this.intervalMs);
    } catch (e) {
      const err = e instanceof Error ? e : new Error(String(e));
      this.backoffMs = this.backoffMs === 0 ? 1000 : Math.min(this.backoffMs * 2, 10000);
      this.opts.onError?.(err, this.backoffMs);
      this.schedule(Math.max(this.intervalMs, this.backoffMs));
    } finally {
      this.inFlight = false;
    }
  }
}
