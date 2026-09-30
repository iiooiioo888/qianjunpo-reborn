#!/usr/bin/env node
/**
 * 靜態檢查 mock 快照與 registry 鍵對齊（無 Cocos）。
 * 在倉庫根目錄：node client/scripts/validate-mock-tactical-display.mjs
 */
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const snapshotPath = join(root, 'assets/resources/data/tactical/demo_initial.json');

const KNOWN_UNIT_TYPES = new Set([0, 2]);
const HUD_CHAR_KEYS = ['char_caocao', 'char_zhangfei', 'char_wu_placeholder'];

function fail(msg) {
  console.error(`validate-mock-tactical-display: FAIL — ${msg}`);
  process.exit(1);
}

function ok(msg) {
  console.log(`validate-mock-tactical-display: OK — ${msg}`);
}

const raw = JSON.parse(readFileSync(snapshotPath, 'utf8'));
if (raw.schemaVersion !== 1 || raw.boardSize !== 19) {
  fail('demo_initial schemaVersion/boardSize');
}

for (const unit of raw.units) {
  if (unit.hp <= 0) continue;
  if (!KNOWN_UNIT_TYPES.has(unit.type)) {
    fail(`unit ${unit.id} type=${unit.type} not in UnitSpriteRegistry mapping`);
  }
  const cell = raw.cells[unit.y]?.[unit.x];
  if (!cell || cell.unitId !== unit.id) {
    fail(`unit ${unit.id} at (${unit.x},${unit.y}) cell.unitId=${cell?.unitId}`);
  }
}

for (let y = 0; y < raw.boardSize; y++) {
  for (let x = 0; x < raw.boardSize; x++) {
    const uid = raw.cells[y][x].unitId;
    if (uid === 0) continue;
    const unit = raw.units.find((u) => u.id === uid && u.hp > 0);
    if (!unit || unit.x !== x || unit.y !== y) {
      fail(`cell (${x},${y}) unitId=${uid} inconsistent with units[]`);
    }
  }
}

const displayDir = join(root, 'assets/scripts/display');
const boardViewSrc = readFileSync(join(displayDir, 'TacticalBoardView.ts'), 'utf8');
const interactionSrc = readFileSync(join(displayDir, 'TacticalBoardInteraction.ts'), 'utf8');
const unitViewSrc = readFileSync(join(displayDir, 'UnitPlaceholderView.ts'), 'utf8');
const visualsSrc = readFileSync(join(displayDir, 'BoardSelectionVisuals.ts'), 'utf8');

if (!boardViewSrc.includes('setSelection(') || !boardViewSrc.includes('BoardSelectionVisuals')) {
  fail('TacticalBoardView missing shared selection visuals');
}
if (!interactionSrc.includes('clearSelection') || !interactionSrc.includes('selectedUnitId === unitAt.id')) {
  fail('TacticalBoardInteraction missing deselect-on-re-tap');
}
if (!unitViewSrc.includes('SelectionRing') || !unitViewSrc.includes('drawUnitSelectionRing')) {
  fail('UnitPlaceholderView missing sprite-safe selection ring');
}
if (!visualsSrc.includes('drawLegalMoveCell')) {
  fail('BoardSelectionVisuals missing legal cell helper');
}

ok(
  `demo_initial units=${raw.units.length}, lockstep=${raw.lockstepFrame}; HUD char keys=${HUD_CHAR_KEYS.join(',')}; selection UX static checks passed`,
);
