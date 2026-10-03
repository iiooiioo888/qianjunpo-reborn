import { ViewSnapshot } from './TacticalSnapshot';

/** 單位 type 與 UnitSpriteRegistry 對齊（0=infantry, 1=archer, 2=cavalry）。 */
export const MOCK_KNOWN_UNIT_TYPES = new Set([0, 1, 2]);

/**
 * 驗證 cells.unitId 與 units 座標一致（mock／live 快照聯調用）。
 * @returns 第一個錯誤描述；一致則 null
 */
export function validateSnapshotCellUnitSync(snap: ViewSnapshot): string | null {
  for (const unit of snap.units) {
    if (unit.hp <= 0) {
      continue;
    }
    const row = snap.cells[unit.y];
    if (!row) {
      return `unit ${unit.id}: missing cells row y=${unit.y}`;
    }
    const cell = row[unit.x];
    if (!cell) {
      return `unit ${unit.id}: missing cell (${unit.x},${unit.y})`;
    }
    if (cell.unitId !== unit.id) {
      return `unit ${unit.id} at (${unit.x},${unit.y}) but cell.unitId=${cell.unitId}`;
    }
  }

  for (let y = 0; y < snap.boardSize; y++) {
    for (let x = 0; x < snap.boardSize; x++) {
      const uid = snap.cells[y][x].unitId;
      if (uid === 0) {
        continue;
      }
      const unit = snap.units.find((u) => u.id === uid && u.hp > 0);
      if (!unit) {
        return `cell (${x},${y}) unitId=${uid} but no live unit`;
      }
      if (unit.x !== x || unit.y !== y) {
        return `cell (${x},${y}) unitId=${uid} but unit at (${unit.x},${unit.y})`;
      }
    }
  }
  return null;
}

export function validateSnapshotUnitTypesForRegistry(snap: ViewSnapshot): string | null {
  for (const unit of snap.units) {
    if (unit.hp <= 0) {
      continue;
    }
    if (!MOCK_KNOWN_UNIT_TYPES.has(unit.type)) {
      return `unit ${unit.id} type=${unit.type} has no UnitSpriteRegistry mapping`;
    }
  }
  return null;
}
