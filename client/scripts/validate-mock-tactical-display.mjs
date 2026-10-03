#!/usr/bin/env node
/**
 * 靜態檢查 mock 快照與 registry 鍵對齊（無 Cocos）。
 * 在倉庫根目錄：node client/scripts/validate-mock-tactical-display.mjs
 */
import { createHash } from 'node:crypto';
import { readFileSync, existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { inflateSync } from 'node:zlib';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const snapshotPath = join(root, 'assets/resources/data/tactical/demo_initial.json');

const KNOWN_UNIT_TYPES = new Set([0, 1, 2]);
const HUD_CHAR_KEYS = ['char_caocao', 'char_zhangfei', 'char_wu_placeholder'];
const CHAR_CARD_V04_STEMS = [
  'PX2D_CHAR_WEI_Caocao_ex_v04',
  'PX2D_CHAR_SHU_Zhangfei_ex_v04',
  'PX2D_CHAR_WU_Placeholder_01_v04',
];
const UNIT_STANDARD_STEMS = ['PX2D_unit_infantry', 'PX2D_unit_archer', 'PX2D_unit_cavalry'];
const ISO25_TILE_STEMS = [
  'ISO25_tile_grass_v02',
  'ISO25_tile_mountain_v01',
  'ISO25_tile_water_v01',
  'ISO25_tile_forest_v01',
];
/** art-manager PASS (#97) — sync 後須與 art/25d/_wip/tiles 一致 */
const ISO25_TILE_MD5_EXPECTED = {
  ISO25_tile_water_v01: '7b5b1c2f057e7f250924a421a10a648d',
  ISO25_tile_forest_v01: '4bec6b8f8ee72e5d24c39e650fd64c0e',
};

function md5File(path) {
  return createHash('md5').update(readFileSync(path)).digest('hex');
}

/** 64×32 RGBA PNG：鑽石外四角須 alpha=0（ISO25 STANDARD #28）。 */
function assertIso25DiamondCornersTransparent(pngPath) {
  const buf = readFileSync(pngPath);
  if (buf.length < 8 || buf.readUInt32BE(0) !== 0x89504e47) {
    fail(`${pngPath} is not a PNG`);
  }
  let offset = 8;
  let width = 0;
  let height = 0;
  let colorType = 0;
  const idatParts = [];
  while (offset + 8 <= buf.length) {
    const len = buf.readUInt32BE(offset);
    const type = buf.toString('ascii', offset + 4, offset + 8);
    const dataStart = offset + 8;
    if (type === 'IHDR' && len >= 13) {
      width = buf.readUInt32BE(dataStart);
      height = buf.readUInt32BE(dataStart + 4);
      colorType = buf[dataStart + 9];
    } else if (type === 'IDAT') {
      idatParts.push(buf.subarray(dataStart, dataStart + len));
    } else if (type === 'IEND') {
      break;
    }
    offset += 12 + len;
  }
  if (width !== 64 || height !== 32) {
    fail(`${pngPath} expected 64x32, got ${width}x${height}`);
  }
  if (colorType !== 6) {
    fail(`${pngPath} expected RGBA color type 6, got ${colorType}`);
  }
  const rawScan = inflateSync(Buffer.concat(idatParts));
  const bpp = 4;
  const rowBytes = width * bpp;
  const stride = 1 + rowBytes;
  const rgba = new Uint8Array(width * height * 4);
  const prevRow = new Uint8Array(rowBytes);
  const curRow = new Uint8Array(rowBytes);

  function paeth(a, b, c) {
    const p = a + b - c;
    const pa = Math.abs(p - a);
    const pb = Math.abs(p - b);
    const pc = Math.abs(p - c);
    if (pa <= pb && pa <= pc) return a;
    if (pb <= pc) return b;
    return c;
  }

  for (let y = 0; y < height; y++) {
    const rowStart = y * stride;
    const filter = rawScan[rowStart];
    for (let i = 0; i < rowBytes; i++) {
      const f = rawScan[rowStart + 1 + i];
      const left = i >= bpp ? curRow[i - bpp] : 0;
      const up = prevRow[i];
      const upLeft = i >= bpp ? prevRow[i - bpp] : 0;
      let raw = f;
      if (filter === 1) raw = (f + left) & 0xff;
      else if (filter === 2) raw = (f + up) & 0xff;
      else if (filter === 3) raw = (f + Math.floor((left + up) / 2)) & 0xff;
      else if (filter === 4) raw = (f + paeth(left, up, upLeft)) & 0xff;
      else if (filter !== 0) {
        fail(`${pngPath} unsupported PNG filter ${filter} at row ${y}`);
      }
      curRow[i] = raw;
    }
    rgba.set(curRow, y * rowBytes);
    prevRow.set(curRow);
  }
  const corners = [
    [0, 0],
    [width - 1, 0],
    [0, height - 1],
    [width - 1, height - 1],
  ];
  for (const [x, y] of corners) {
    const a = rgba[(y * width + x) * 4 + 3];
    if (a !== 0) {
      fail(`${pngPath} corner (${x},${y}) alpha=${a}, expected 0`);
    }
  }
}

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
if (
  !lockstepHudSrc.includes('formatSelectionLine') ||
  !hudSrc.includes('setSelectedUnitId') ||
  !hudSrc.includes('formatSelectionLine')
) {
  fail('TimeFlowHudStub/LockstepHudFormat missing selected unit grid HUD line');
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
const tileRegistrySrc = readFileSync(join(displayDir, 'TerrainTileSpriteRegistry.ts'), 'utf8');
for (const stem of ISO25_TILE_STEMS) {
  if (!tileRegistrySrc.includes(`'${stem}'`)) {
    fail(`TerrainTileSpriteRegistry missing default stem ${stem}`);
  }
}
if (tileRegistrySrc.includes('grass_v01')) {
  fail('TerrainTileSpriteRegistry still references rejected grass_v01');
}
const tilesDir = join(root, 'assets/resources/textures/2d/tiles');
const previewTilesDir = join(root, 'static-preview/tiles');
for (const stem of ISO25_TILE_STEMS) {
  const png = join(tilesDir, `${stem}.png`);
  if (!existsSync(png)) {
    fail(`missing ${png} — run bash client/scripts/sync-wip-tile-textures.sh`);
  }
  const previewPng = join(previewTilesDir, `${stem}.png`);
  if (!existsSync(previewPng)) {
    fail(`missing ${previewPng} — run bash client/scripts/sync-wip-tile-textures.sh`);
  }
  const expectedMd5 = ISO25_TILE_MD5_EXPECTED[stem];
  if (expectedMd5) {
    const cocosMd5 = md5File(png);
    const previewMd5 = md5File(previewPng);
    if (cocosMd5 !== expectedMd5) {
      fail(`${png} md5 ${cocosMd5} !== expected ${expectedMd5}`);
    }
    if (previewMd5 !== expectedMd5) {
      fail(`${previewPng} md5 ${previewMd5} !== expected ${expectedMd5}`);
    }
    assertIso25DiamondCornersTransparent(png);
    assertIso25DiamondCornersTransparent(previewPng);
  }
}
const staticPreviewConfig = readFileSync(join(root, 'static-preview/config.js'), 'utf8');
const staticPreviewAssetRegistry = readFileSync(join(root, 'static-preview/asset-registry.js'), 'utf8');
for (const stem of ISO25_TILE_STEMS) {
  if (!staticPreviewAssetRegistry.includes(stem)) {
    fail(`static-preview/asset-registry.js missing tile stem ${stem}`);
  }
}
const staticPreviewBoard = readFileSync(join(root, 'static-preview/board.js'), 'utf8');
if (!staticPreviewBoard.includes('drawTerrainTile')) {
  fail('static-preview/board.js missing ISO25 terrain tile draw path');
}
if (!staticPreviewBoard.includes('boardIsoTileSourceRect')) {
  fail('static-preview/board.js must crop ISO25 opaque diamond before drawImage');
}

const staticPreviewApp = readFileSync(join(root, 'static-preview/app.js'), 'utf8');
const staticPreviewHud = readFileSync(join(root, 'static-preview/hud.js'), 'utf8');
if (
  !staticPreviewApp.includes('pendingCommandFailure') ||
  !staticPreviewApp.includes('LIVE_LABEL_COMMAND_RETRY')
) {
  fail('static-preview missing Live command failure retry UX');
}
if (
  !staticPreviewHud.includes('formatSelectionLine') ||
  !staticPreviewApp.includes('hud-selection') ||
  !staticPreviewApp.includes('formatSelectionLine')
) {
  fail('static-preview missing selected unit grid HUD line');
}
if (!bootstrapSrc.includes('setSelectedUnitId')) {
  fail('TacticalBootstrap missing HUD selection coord wiring');
}
const liveGatewayPreview = readFileSync(
  join(root, 'static-preview/live-gateway.js'),
  'utf8',
);
if (
  !liveGatewayPreview.includes('v1/tactical/step-lockstep') ||
  !staticPreviewApp.includes('stepTacticalLockstep') ||
  !staticPreviewApp.includes('applyLiveLockstepAfterCommand')
) {
  fail('static-preview missing Live step-lockstep after accepted command');
}
if (
  !staticPreviewHud.includes('formatLastSkillCastLine') ||
  !staticPreviewApp.includes('hud-skill-cast') ||
  !staticPreviewApp.includes('formatLastSkillCastLine')
) {
  fail('static-preview missing lastSkillCast HUD line');
}
const battleEndPreview = readFileSync(join(root, 'static-preview/battle-end.js'), 'utf8');
if (
  !battleEndPreview.includes('endReason') ||
  !staticPreviewApp.includes('updateBattleEndOverlay') ||
  !staticPreviewApp.includes('isBattleFinished')
) {
  fail('static-preview missing battle end overlay (winner/endReason)');
}
const battleEndIngest = readFileSync(
  join(root, 'static-preview/battle-end-ingest.js'),
  'utf8',
);
if (
  !battleEndIngest.includes('rememberTerminalOutcome') ||
  !staticPreviewApp.includes('snapshotForBattleEndOverlay') ||
  !staticPreviewApp.includes('stickyTerminalOutcome')
) {
  fail('static-preview missing sticky terminal battle-end ingest (Live wipeout)');
}
if (
  !staticPreviewApp.includes('ensureFreshLiveZoneBeforeEnter') ||
  !readFileSync(join(root, 'static-preview/live-zone-entry.js'), 'utf8').includes(
    'isTerminalLiveSnapshot',
  )
) {
  fail('static-preview missing live zone fresh entry (ended battle → new zoneId)');
}
const staticPreviewIndex = readFileSync(join(root, 'static-preview/index.html'), 'utf8');
if (!staticPreviewIndex.includes("base.href = '/qjp/'")) {
  fail('static-preview index.html missing /qjp/ base bootstrap');
}
if (
  !staticPreviewApp.includes('TACTICAL_COMMAND_KIND_SKILL') ||
  !staticPreviewApp.includes('skill_id') ||
  !staticPreviewApp.includes('hud-cast-skill') ||
  !staticPreviewApp.includes('runLiveSkillCommand')
) {
  fail('static-preview missing Live stub skill cast (kind=5 + skill_id)');
}
const staticPreviewStubSkill = readFileSync(
  join(root, 'static-preview/stub-skill.js'),
  'utf8',
);
if (
  !staticPreviewStubSkill.includes('STUB_SKILL_STRIKE_ID') ||
  !staticPreviewConfig.includes('stubSkillId')
) {
  fail('static-preview stub skill config/helpers missing');
}

const localUnit = raw.units.find((u) => u.owner === 0 && u.hp > 0);
if (!localUnit || localUnit.id !== 101) {
  fail('demo_initial expected local player unit id=101 for char_caocao mapping stub');
}

ok(
  `demo_initial units=${raw.units.length}, lockstep=${raw.lockstepFrame}; unit STANDARD stems+PNGs ok; ISO25 tile stems+PNGs ok; HUD char keys=${HUD_CHAR_KEYS.join(',')}; char v04 stems+PNGs ok; lockstep sync + char-card selection link static checks passed`,
);
