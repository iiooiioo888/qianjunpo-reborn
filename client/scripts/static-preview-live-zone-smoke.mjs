#!/usr/bin/env node
/**
 * Offline checks for live zone rotation before enter-battle.
 * Usage: node client/scripts/static-preview-live-zone-smoke.mjs
 */
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import {
  battleIdForZone,
  isTerminalLiveSnapshot,
  mintFreshLiveZoneId,
} from '../static-preview/live-zone-entry.js';
import { isBattleFinished } from '../static-preview/battle-end.js';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');

function fail(msg) {
  console.error(`FAIL: ${msg}`);
  process.exit(1);
}

function ok(msg) {
  console.log(`ok: ${msg}`);
}

const wipeout = {
  schemaVersion: 1,
  boardSize: 19,
  endReason: 'wipeout',
  winner: 0,
  lockstepFrame: 51,
  cells: Array.from({ length: 19 }, () => Array(19).fill({ unitId: 0 })),
  units: [],
};

if (!isTerminalLiveSnapshot(wipeout)) {
  fail('wipeout snapshot should be terminal');
}
if (isTerminalLiveSnapshot({ schemaVersion: 1, endReason: 'none', winner: null })) {
  fail('in-progress snapshot should not be terminal');
}
if (!isBattleFinished(wipeout)) {
  fail('wipeout should satisfy isBattleFinished');
}

const z = mintFreshLiveZoneId();
if (!/^qjp-[a-z0-9]+-[a-z0-9]+$/.test(z)) {
  fail(`unexpected zone id shape: ${z}`);
}
if (battleIdForZone('default', 0) !== 'default/0') {
  fail('battleIdForZone default/0');
}

const appJs = readFileSync(join(root, 'static-preview/app.js'), 'utf8');
if (!appJs.includes('ensureFreshLiveZoneBeforeEnter')) {
  fail('app.js missing ensureFreshLiveZoneBeforeEnter');
}
if (!appJs.includes('bounceLiveToDeployAfterTerminalZone')) {
  fail('app.js missing deploy bounce after terminal zone');
}

const campaign = readFileSync(join(root, 'static-preview/campaign-flow.js'), 'utf8');
if (!campaign.includes('toDeploy')) {
  fail('campaign-flow missing toDeploy');
}

ok('live zone entry terminal detection + app wiring checks passed');
