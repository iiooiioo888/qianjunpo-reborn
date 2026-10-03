import { manhattan } from './field-iso.js';

const MOVE_RANGE = 4;
const ATTACK_RANGE = 2;

let nextId = 1;

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
    selectedId: null,
    highlightCells: [],
    message: '請選取己方單位。',
  };
}

export function unitAt(state, gx, gy) {
  return state.units.find(
    (u) => u.troops > 0 && u.x === gx && u.y === gy,
  );
}

export function getUnit(state, id) {
  return state.units.find((u) => u.id === id);
}

export function selectUnit(state, unitId) {
  const u = getUnit(state, unitId);
  if (!u || u.side !== 'player' || u.troops <= 0) {
    return { ...state, selectedId: null, highlightCells: [], message: '請選取己方單位。' };
  }
  const highlights = computeMoveHighlights(state, u);
  return {
    ...state,
    selectedId: unitId,
    highlightCells: highlights,
    message: `已選 ${typeName(u.type)}（${u.troops}）· 點空格移動或點敵軍攻擊。`,
  };
}

function typeName(type) {
  return { infantry: '步兵', archer: '弓兵', cavalry: '騎兵' }[type] ?? type;
}

function computeMoveHighlights(state, unit) {
  const cells = [];
  for (let gy = 0; gy < 13; gy++) {
    for (let gx = 0; gx < 13; gx++) {
      if (gx === unit.x && gy === unit.y) continue;
      if (unitAt(state, gx, gy)) continue;
      if (manhattan({ x: gx, y: gy }, unit) <= MOVE_RANGE) {
        cells.push({ x: gx, y: gy, kind: 'move' });
      }
    }
  }
  return cells;
}

export function handleCellClick(state, gx, gy) {
  const clicked = unitAt(state, gx, gy);

  if (!state.selectedId) {
    if (clicked?.side === 'player') {
      return selectUnit(state, clicked.id);
    }
    return { ...state, message: '請先點選己方一支部隊。' };
  }

  const selected = getUnit(state, state.selectedId);
  if (!selected || selected.troops <= 0) {
    return selectUnit(state, null);
  }

  if (clicked?.id === selected.id) {
    return { ...state, selectedId: null, highlightCells: [], message: '已取消選取。' };
  }

  if (clicked?.side === 'player') {
    return selectUnit(state, clicked.id);
  }

  if (clicked?.side === 'enemy') {
    const dist = manhattan(clicked, selected);
    if (dist > ATTACK_RANGE) {
      return { ...state, message: '距離太遠，無法攻擊該敵軍。' };
    }
    const damage = Math.max(8, Math.floor(selected.troops * 0.25));
    const units = state.units.map((u) =>
      u.id === clicked.id
        ? { ...u, troops: Math.max(0, u.troops - damage) }
        : u,
    );
    return {
      ...state,
      units,
      selectedId: null,
      highlightCells: [],
      message: `攻擊 ${typeName(clicked.type)}，敵損 ${damage}，剩 ${Math.max(0, clicked.troops - damage)}。`,
    };
  }

  const canMove = state.highlightCells.some((c) => c.x === gx && c.y === gy);
  if (!canMove) {
    return { ...state, message: '該格不可移動。' };
  }

  const units = state.units.map((u) =>
    u.id === selected.id ? { ...u, x: gx, y: gy } : u,
  );
  const moved = { ...selected, x: gx, y: gy };
  const next = {
    ...state,
    units,
    selectedId: moved.id,
    highlightCells: computeMoveHighlights({ ...state, units }, moved),
    message: `已移動至 (${gx},${gy})。`,
  };
  return next;
}
