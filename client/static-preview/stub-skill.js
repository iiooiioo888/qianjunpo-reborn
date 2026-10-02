/** Aligned with pkg/combat.SkillStubStrike and tactical.KindSkill (5). */
export const STUB_SKILL_STRIKE_ID = 1;
export const TACTICAL_COMMAND_KIND_SKILL = 5;

/** Unit type → melee/ranged range (pkg/combat/default_combat.json). */
const ATTACK_RANGE_BY_UNIT_TYPE = {
  0: 1,
  1: 3,
  2: 1,
};

export function attackRangeForUnitType(unitType) {
  return ATTACK_RANGE_BY_UNIT_TYPE[unitType] ?? 1;
}

export function chebyshevDistance(ax, ay, bx, by) {
  return Math.max(Math.abs(ax - bx), Math.abs(ay - by));
}

export function isInAttackRange(attacker, targetX, targetY) {
  const range = attackRangeForUnitType(attacker.type);
  const d = chebyshevDistance(attacker.x, attacker.y, targetX, targetY);
  return d >= 1 && d <= range;
}

/**
 * First living enemy in attack range of `caster` (stable by unit id).
 * @param {import('./types.js').ViewSnapshot} snap
 */
export function findStubStrikeTarget(snap, caster) {
  const enemies = snap.units
    .filter((u) => u.owner !== caster.owner && u.hp > 0)
    .sort((a, b) => a.id - b.id);
  for (const enemy of enemies) {
    if (isInAttackRange(caster, enemy.x, enemy.y)) {
      return { x: enemy.x, y: enemy.y, unitId: enemy.id };
    }
  }
  return null;
}
