import { ViewSnapshot } from '../logic/TacticalSnapshot';

export type HudSyncSource = 'mock' | 'live';

export type HudSyncLinkState = 'connected' | 'stale' | 'error';

/** Display-only lockstep / Janus sync context (not authoritative). */
export interface HudLockstepSyncContext {
  source: HudSyncSource;
  link: HudSyncLinkState;
  /** Live: ms since last successful Janus snapshot apply. */
  pollAgeMs?: number;
  /** Live: configured poll interval for stale threshold hints. */
  pollIntervalMs?: number;
}

function shortenHash(hash: string): string {
  if (hash.length <= 12) {
    return hash;
  }
  return `${hash.slice(0, 10)}…`;
}

/** Primary lockstep HUD line: frame from ViewSnapshot + sync mode string. */
export function formatLockstepFrameLine(snap: ViewSnapshot, ctx: HudLockstepSyncContext): string {
  const hash = shortenHash(snap.initialStateHash);
  if (ctx.source === 'mock') {
    const link =
      ctx.link === 'error'
        ? 'load/sync error'
        : ctx.link === 'stale'
          ? 'stale (local drift)'
          : 'local JSON · non-authoritative';
    return `lockstep frame: ${snap.lockstepFrame}  ·  sync: mock · ${link}  ·  hash ${hash}`;
  }

  const age =
    ctx.pollAgeMs != null && ctx.pollAgeMs >= 0 ? ` · poll age ${Math.round(ctx.pollAgeMs)}ms` : '';
  const interval =
    ctx.pollIntervalMs != null && ctx.pollIntervalMs > 0 ? ` / ${ctx.pollIntervalMs}ms` : '';
  const link =
    ctx.link === 'error'
      ? 'Janus error'
      : ctx.link === 'stale'
        ? `stale${interval}${age}`
        : `Janus mirror · fresh${age}`;
  return `lockstep frame: ${snap.lockstepFrame}  ·  sync: live · ${link}  ·  hash ${hash}`;
}

export function mockSyncContextFromBootstrap(link: HudSyncLinkState = 'connected'): HudLockstepSyncContext {
  return { source: 'mock', link };
}

export function liveSyncContextFromPoll(
  pollAgeMs: number,
  pollIntervalMs: number,
  link: HudSyncLinkState,
): HudLockstepSyncContext {
  return { source: 'live', link, pollAgeMs, pollIntervalMs };
}

/** Stale when live poll age exceeds ~2× configured interval. */
export function liveLinkStateFromPollAge(pollAgeMs: number, pollIntervalMs: number): HudSyncLinkState {
  const threshold = Math.max(pollIntervalMs * 2, pollIntervalMs + 500);
  return pollAgeMs > threshold ? 'stale' : 'connected';
}
