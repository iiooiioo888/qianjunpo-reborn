#!/usr/bin/env node
/**
 * 靜態檢查 mock 快照與 registry 鍵對齊（無 Cocos）。
 * 在倉庫根目錄：node client/scripts/validate-mock-tactical-display.mjs
 */
import { readFileSync, existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const snapshotPath = join(root, 'assets/resources/data/tactical/demo_initial.json');

const KNOWN_UNIT_TYPES = new Set([0, 2]);
const HUD_CHAR_KEYS = ['char_caocao', 'char_zhangfei', 'char_wu_placeholder'];
const CHAR_CARD_V04_STEMS = [
  'PX2D_CHAR_WEI_Caocao_ex_v04',
  'PX2D_CHAR_SHU_Zhangfei_ex_v04',
  'PX2D_CHAR_WU_Placeholder_01_v04',
];
const UNIT_STANDARD_STEMS = ['PX2D_unit_infantry', 'PX2D_unit_cavalry'];

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
const hudSrc = readFileSync(join(displayDir, 'TimeFlowHudStub.ts'), 'utf8');
const lockstepHudSrc = readFileSync(join(displayDir, 'LockstepHudFormat.ts'), 'utf8');
const charStripSrc = readFileSync(join(displayDir, 'CharacterCardHudStrip.ts'), 'utf8');
const charMapSrc = readFileSync(join(displayDir, 'UnitCharacterCardMapping.ts'), 'utf8');
const bootstrapSrc = readFileSync(join(root, 'assets/scripts/app/TacticalBootstrap.ts'), 'utf8');

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
if (
  !hudSrc.includes('formatLockstepFrameLine') ||
  !hudSrc.includes('setLockstepSyncContext') ||
  !lockstepHudSrc.includes('sync: mock')
) {
  fail('TimeFlowHudStub missing lockstep sync HUD line (LockstepHudFormat)');
}

const unitRegistrySrc = readFileSync(join(displayDir, 'UnitSpriteRegistry.ts'), 'utf8');
for (const stem of UNIT_STANDARD_STEMS) {
  if (!unitRegistrySrc.includes(`'${stem}'`)) {
    fail(`UnitSpriteRegistry missing default stem ${stem}`);
  }
}
if (unitRegistrySrc.includes('_v02')) {
  fail('UnitSpriteRegistry still references _v02 unit stems');
}
const unitsDir = join(root, 'assets/resources/textures/2d/units');
for (const stem of UNIT_STANDARD_STEMS) {
  const png = join(unitsDir, `${stem}.png`);
  if (!existsSync(png)) {
    fail(`missing ${png} — run bash client/scripts/sync-wip-unit-textures.sh`);
  }
}

const registrySrc = readFileSync(join(displayDir, 'CharacterCardSpriteRegistry.ts'), 'utf8');
for (const stem of CHAR_CARD_V04_STEMS) {
  if (!registrySrc.includes(`'${stem}'`)) {
    fail(`CharacterCardSpriteRegistry missing default stem ${stem}`);
  }
}
if (registrySrc.includes('_v02')) {
  fail('CharacterCardSpriteRegistry still references _v02 character stems');
}
const charsDir = join(root, 'assets/resources/textures/2d/chars');
for (const stem of CHAR_CARD_V04_STEMS) {
  const png = join(charsDir, `${stem}.png`);
  if (!existsSync(png)) {
    fail(`missing ${png} — run bash client/scripts/sync-wip-char-textures.sh`);
  }
}

if (!charStripSrc.includes('setSelectionLinkedCard') || !visualsSrc.includes('drawSelectedCharCardFrame')) {
  fail('CharacterCardHudStrip missing selection-linked card highlight');
}
if (
  !charMapSrc.includes('resolveCharCardKeyForUnit') ||
  !charMapSrc.includes('char_caocao') ||
  !charMapSrc.includes('101')
) {
  fail('UnitCharacterCardMapping missing mock unitId/type → char card stub');
}
if (
  !interactionSrc.includes('onSelectionChange') ||
  !bootstrapSrc.includes('syncCharCardHighlight') ||
  !bootstrapSrc.includes('resolveCharCardKeyForUnit')
) {
  fail('TacticalBootstrap/Interaction missing selection → char card strip wiring');
}
if (
  !bootstrapSrc.includes('prepareJanusLiveSession') ||
  !bootstrapSrc.includes('useLiveJanus')
) {
  fail('TacticalBootstrap missing Live HTTP connect→enter-battle prepare');
}
const liveGatewaySrc = readFileSync(
  join(root, 'assets/scripts/network/JanusLiveGatewayHttp.ts'),
  'utf8',
);
if (
  !liveGatewaySrc.includes('v1/tactical/connect') ||
  !liveGatewaySrc.includes('v1/tactical/enter-battle')
) {
  fail('JanusLiveGatewayHttp missing #60 tactical connect/enter-battle paths');
}
if (
  !hudSrc.includes('liveCommandRetry') ||
  !hudSrc.includes('重試戰術指令') ||
  !bootstrapSrc.includes('liveCommandRetry') ||
  !bootstrapSrc.includes('pendingLiveCommandError')
) {
  fail('TimeFlowHudStub/TacticalBootstrap missing Live tactical command retry HUD');
}
const staticPreviewApp = readFileSync(join(root, 'static-preview/app.js'), 'utf8');
if (
  !staticPreviewApp.includes('pendingCommandFailure') ||
  !staticPreviewApp.includes('LIVE_LABEL_COMMAND_RETRY')
) {
  fail('static-preview missing Live command failure retry UX');
}

const localUnit = raw.units.find((u) => u.owner === 0 && u.hp > 0);
if (!localUnit || localUnit.id !== 101) {
  fail('demo_initial expected local player unit id=101 for char_caocao mapping stub');
}

ok(
  `demo_initial units=${raw.units.length}, lockstep=${raw.lockstepFrame}; unit STANDARD stems+PNGs ok; HUD char keys=${HUD_CHAR_KEYS.join(',')}; char v04 stems+PNGs ok; lockstep sync + char-card selection link static checks passed`,
);
