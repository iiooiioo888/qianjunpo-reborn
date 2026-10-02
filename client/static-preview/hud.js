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

/** Grid coords for the locally selected unit (display-only). */
export function formatSelectionLine(snap, selectedUnitId) {
  if (selectedUnitId == null || !snap) {
    return 'selected unit: —';
  }
  const unit = snap.units.find((u) => u.id === selectedUnitId && u.hp > 0);
  if (!unit) {
    return `selected unit: id=${selectedUnitId}  ·  grid (—,—)`;
  }
  return `selected unit: id=${unit.id}  ·  grid (${unit.x}, ${unit.y})`;
}

function normalizeSkillCast(cast) {
  if (!cast || typeof cast !== 'object') {
    return null;
  }
  const skillId = cast.skillId ?? cast.skill_id;
  const caster = cast.casterUnitId ?? cast.caster_unit_id;
  const tx = cast.targetX ?? cast.target_x;
  const ty = cast.targetY ?? cast.target_y;
  const frame = cast.lockstepFrame ?? cast.lockstep_frame;
  if (skillId == null && caster == null && tx == null && ty == null) {
    return null;
  }
  return { skillId, caster, tx, ty, frame };
}

/** Pick latest cast from snapshot or sticky client cache (avoids HUD blanking between polls). */
export function pickLastSkillCast(snap, stickyCast) {
  return normalizeSkillCast(snap?.lastSkillCast) ?? normalizeSkillCast(stickyCast);
}

/** Latest KindSkill resolution from authoritative snapshot (display-only). */
export function formatLastSkillCastLine(snap, stickyCast = null) {
  const cast = pickLastSkillCast(snap, stickyCast);
  if (!cast) {
    return 'lastSkillCast: —';
  }
  const frame = cast.frame ?? '?';
  const stickyNote = snap?.lastSkillCast ? '' : ' · (sticky)';
  return `lastSkillCast: skill=${cast.skillId ?? '?'}  caster=${cast.caster ?? '?'}  → (${cast.tx ?? '?'},${cast.ty ?? '?'})  @ frame ${frame}${stickyNote}`;
}

export function lastSkillCastHudClass(snap, stickyCast = null) {
  return pickLastSkillCast(snap, stickyCast) ? 'skill-cast-active' : '';
}
