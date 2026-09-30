export function timeFlowRateToFloat(parts) {
  return parts / 10000;
}

function shortenHash(hash) {
  if (!hash || hash.length <= 12) {
    return hash ?? '';
  }
  return `${hash.slice(0, 10)}…`;
}

export function formatRateLine(snap) {
  const rate = timeFlowRateToFloat(snap.timeFlowRateParts);
  return `timeFlowRateParts: ${snap.timeFlowRateParts}  (${rate.toFixed(2)}×)`;
}

export function formatLockstepFrameLine(snap, ctx) {
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

export function liveLinkStateFromPollAge(pollAgeMs, pollIntervalMs) {
  const threshold = Math.max(pollIntervalMs * 2, pollIntervalMs + 500);
  return pollAgeMs > threshold ? 'stale' : 'connected';
}
