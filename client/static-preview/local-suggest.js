import { LOCAL_PLAYER_OWNER } from './config.js';
import { computeLegalMoveDestinations } from './reachability.js';
import {
  findStubStrikeTarget,
  STUB_SKILL_STRIKE_ID,
  TACTICAL_COMMAND_KIND_SKILL,
} from './stub-skill.js';

export const SUGGESTION_KIND_MOVE = 'move';
export const SUGGESTION_KIND_SKILL = 'skill';

function chebyshev(ax, ay, bx, by) {
  return Math.max(Math.abs(ax - bx), Math.abs(ay - by));
}

function nearestEnemy(snap, unit) {
  const enemies = snap.units
    .filter((u) => u.owner !== unit.owner && u.hp > 0)
    .sort((a, b) => a.id - b.id);
  if (enemies.length === 0) {
    return null;
  }
  let best = enemies[0];
  let bestD = chebyshev(unit.x, unit.y, best.x, best.y);
  for (const e of enemies.slice(1)) {
    const d = chebyshev(unit.x, unit.y, e.x, e.y);
    if (d < bestD) {
      bestD = d;
      best = e;
    }
  }
  return best;
}

/**
 * Deterministic client AI when POST /v1/suggest is unavailable (mirrors pkg/ai intent: strike or close).
 * @returns {{ kind: string, unitId: number, to: { x: number, y: number }, skillId?: number, tacticalKind: number }}
 */
export function deriveLocalSuggestion(snap) {
  const friendlies = snap.units
    .filter((u) => u.owner === LOCAL_PLAYER_OWNER && u.hp > 0)
    .sort((a, b) => a.id - b.id);

  for (const unit of friendlies) {
    const strike = findStubStrikeTarget(snap, unit);
    if (strike) {
      return {
        kind: SUGGESTION_KIND_SKILL,
        unitId: unit.id,
        to: { x: strike.x, y: strike.y },
        skillId: STUB_SKILL_STRIKE_ID,
        tacticalKind: TACTICAL_COMMAND_KIND_SKILL,
        inferPath: 'local',
      };
    }
  }

  let bestMove = null;
  for (const unit of friendlies) {
    const enemy = nearestEnemy(snap, unit);
    if (!enemy) {
      continue;
    }
    const legal = computeLegalMoveDestinations(snap, unit.id);
    for (const dest of legal) {
      if (dest.x === unit.x && dest.y === unit.y) {
        continue;
      }
      const d = chebyshev(dest.x, dest.y, enemy.x, enemy.y);
      if (!bestMove || d < bestMove.score) {
        bestMove = { unitId: unit.id, to: dest, score: d };
      }
    }
  }

  if (bestMove) {
    return {
      kind: SUGGESTION_KIND_MOVE,
      unitId: bestMove.unitId,
      to: bestMove.to,
      tacticalKind: 1,
      inferPath: 'local',
    };
  }

  const fallback = friendlies[0];
  if (fallback) {
    return {
      kind: SUGGESTION_KIND_MOVE,
      unitId: fallback.id,
      to: { x: fallback.x, y: fallback.y },
      tacticalKind: 1,
      inferPath: 'local',
    };
  }

  return null;
}

/**
 * @param {object} raw from infer HTTP
 */
export function normalizeInferSuggestion(raw, snap) {
  if (!raw || typeof raw !== 'object') {
    return null;
  }
  const kind = raw.kind;
  const unitId = Number(raw.unit_id ?? raw.unitId);
  const toX = Number(raw.to_x ?? raw.toX);
  const toY = Number(raw.to_y ?? raw.toY);
  if (!Number.isFinite(unitId) || !Number.isFinite(toX) || !Number.isFinite(toY)) {
    return null;
  }
  const unit = snap?.units?.find((u) => u.id === unitId && u.hp > 0);
  if (!unit || unit.owner !== LOCAL_PLAYER_OWNER) {
    return null;
  }
  if (kind === SUGGESTION_KIND_SKILL) {
    const skillId = Number(raw.skill_id ?? raw.skillId ?? STUB_SKILL_STRIKE_ID);
    return {
      kind: SUGGESTION_KIND_SKILL,
      unitId,
      to: { x: toX, y: toY },
      skillId,
      tacticalKind: TACTICAL_COMMAND_KIND_SKILL,
      inferPath: raw.infer_path ?? raw.inferPath ?? 'infer',
    };
  }
  if (kind === SUGGESTION_KIND_MOVE || kind == null) {
    return {
      kind: SUGGESTION_KIND_MOVE,
      unitId,
      to: { x: toX, y: toY },
      tacticalKind: 1,
      inferPath: raw.infer_path ?? raw.inferPath ?? 'infer',
    };
  }
  return null;
}

export function summarizeBattleContext(snap) {
  const alive = snap.units.filter((u) => u.hp > 0);
  const p0 = alive.filter((u) => u.owner === LOCAL_PLAYER_OWNER).length;
  const p1 = alive.filter((u) => u.owner !== LOCAL_PLAYER_OWNER).length;
  return {
    units_summary: `owner0=${p0} owner1=${p1} frame=${snap.lockstepFrame}`,
    terrain_summary: 'iso25 tactical board',
  };
}
