import { manhattan } from './field-iso.js';

const MOVE_RANGE = 4;
const ATTACK_RANGE = 2;
const GRID_SIZE = 13;

let nextId = 1;

export function calcTroopDamage(attackerTroops) {
  return Math.max(8, Math.floor(attackerTroops * 0.25));
}

export function createBattleFromDeploy(deploy) {
  nextId = 1;
  const units = [];

  const playerSlots = [
    { x: 2, y: 4 },
    { x: 2, y: 6 },
    { x: 2, y: 8 },
  ];
  const types = ['infantry', 'archer', 'cavalry'];
  types.forEach((type, i) => {
    units.push({
      id: nextId++,
      side: 'player',
      type,
      troops: deploy[type],
      x: playerSlots[i].x,
      y: playerSlots[i].y,
    });
  });

  units.push({
    id: nextId++,
    side: 'enemy',
    type: 'infantry',
    troops: 150,
    x: 10,
    y: 5,
  });
  units.push({
    id: nextId++,
    side: 'enemy',
    type: 'archer',
    troops: 120,
    x: 10,
    y: 8,
  });

  return {
    units,
    turnNumber: 0,
    popups: [],
    message: '自動戰鬥進行中…',
    gameResult: null,
  };
}

function sumSideTroops(state, side) {
  return state.units
    .filter((u) => u.side === side)
    .reduce((sum, u) => sum + u.troops, 0);
}

function resolveGameResult(state) {
  if (sumSideTroops(state, 'player') <= 0) return 'lose';
  if (sumSideTroops(state, 'enemy') <= 0) return 'win';
  return null;
}

function applyEndGame(state) {
  const gameResult = resolveGameResult(state);
  if (!gameResult) return state;
  return {
    ...state,
    gameResult,
    message:
      gameResult === 'win' ? '勝利！敵軍全滅。' : '失敗！己方全滅。',
  };
}

export function unitAt(state, gx, gy) {
  return state.units.find(
    (u) => u.troops > 0 && u.x === gx && u.y === gy,
  );
}

function typeName(type) {
  return { infantry: '步兵', archer: '弓兵', cavalry: '騎兵' }[type] ?? type;
}

function livingUnits(state) {
  return state.units.filter((u) => u.troops > 0);
}

function findNearestEnemy(unit, state) {
  const enemies = livingUnits(state).filter((u) => u.side !== unit.side);
  let best = null;
  let bestDist = Infinity;
  for (const e of enemies) {
    const d = manhattan(unit, e);
    if (d < bestDist) {
      bestDist = d;
      best = e;
    }
  }
  return { target: best, dist: bestDist };
}

function bestMoveCell(unit, target, state) {
  let best = null;
  let bestDist = manhattan(unit, target);
  for (let gy = 0; gy < GRID_SIZE; gy++) {
    for (let gx = 0; gx < GRID_SIZE; gx++) {
      if (gx === unit.x && gy === unit.y) continue;
      if (unitAt(state, gx, gy)) continue;
      const step = manhattan({ x: gx, y: gy }, unit);
      if (step > MOVE_RANGE || step === 0) continue;
      const toTarget = manhattan({ x: gx, y: gy }, target);
      if (toTarget < bestDist) {
        bestDist = toTarget;
        best = { x: gx, y: gy };
      }
    }
  }
  return best;
}

function appendPopup(state, gx, gy, text) {
  const popups = [
    ...(state.popups ?? []),
    { gx, gy, text, ttl: 48 },
  ];
  return { ...state, popups };
}

function applyAttack(state, attacker, defender) {
  const damage = calcTroopDamage(attacker.troops);
  const remaining = Math.max(0, defender.troops - damage);
  const units = state.units.map((u) =>
    u.id === defender.id ? { ...u, troops: remaining } : u,
  );
  let next = { ...state, units };
  next = appendPopup(
    next,
    defender.x,
    defender.y,
    `-${damage}`,
  );
  return {
    ...next,
    message: `${sideLabel(attacker.side)}${typeName(attacker.type)} 攻擊 ${typeName(defender.type)}，敵損 ${damage}。`,
  };
}

function sideLabel(side) {
  return side === 'player' ? '己方' : '敵方';
}

function moveUnit(state, unitId, gx, gy) {
  const units = state.units.map((u) =>
    u.id === unitId ? { ...u, x: gx, y: gy } : u,
  );
  return { ...state, units };
}

/** 單一單位本回合行動：射程內攻擊，否則朝最近敵軍移動一格距內最佳格 */
function actUnit(state, unit) {
  if (unit.troops <= 0) return state;
  const { target, dist } = findNearestEnemy(unit, state);
  if (!target) return state;

  if (dist <= ATTACK_RANGE) {
    return applyAttack(state, unit, target);
  }

  const dest = bestMoveCell(unit, target, state);
  if (!dest) return state;
  const moved = moveUnit(state, unit.id, dest.x, dest.y);
  return {
    ...moved,
    message: `${sideLabel(unit.side)}${typeName(unit.type)} 向敵軍靠近。`,
  };
}

function turnOrderUnits(state) {
  return livingUnits(state).sort((a, b) => {
    if (a.side !== b.side) return a.side === 'player' ? -1 : 1;
    return a.id - b.id;
  });
}

/**
 * 推進一個完整回合：己方各堆行動後，敵方各堆再行動（含反擊移動／攻擊）。
 */
export function advanceAutoTurn(state) {
  if (state.gameResult) return state;

  let next = { ...state, turnNumber: (state.turnNumber ?? 0) + 1 };
  const order = turnOrderUnits(next);

  for (const snapshot of order) {
    const unit = next.units.find((u) => u.id === snapshot.id);
    if (!unit || unit.troops <= 0) continue;

    next = actUnit(next, unit);
    next = applyEndGame(next);
    if (next.gameResult) break;
  }

  if (!next.gameResult) {
    next = {
      ...next,
      message: `第 ${next.turnNumber} 回合 · 己方 ${sumSideTroops(next, 'player')} / 敵方 ${sumSideTroops(next, 'enemy')}`,
    };
  }

  return next;
}

/** 動畫帧：衰減飄字與技能閃光 */
export function tickBattleFx(state) {
  if (!state) return state;
  const popups = (state.popups ?? [])
    .map((p) => ({ ...p, ttl: p.ttl - 1 }))
    .filter((p) => p.ttl > 0);
  return { ...state, popups };
}

/** 無 UI 快速模擬至結束（除錯／自測） */
export function simulateBattleToEnd(deploy, maxTurns = 500) {
  let state = createBattleFromDeploy(deploy);
  let turns = 0;
  while (!state.gameResult && turns < maxTurns) {
    state = advanceAutoTurn(state);
    turns += 1;
  }
  return { state, turns, log: state.message };
}
