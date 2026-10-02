#!/usr/bin/env node
/**
 * Offline checks for battle-end overlay contract (wipeout / mockVictory).
 * Live HTTP wipeout still requires core: WIPEOUT_SMOKE=1 ./scripts/janus-http-smoke.sh
 *
 * Usage (from repo root):
 *   node client/scripts/static-preview-battle-end-smoke.mjs
 */
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import {
  applyMockVictoryPatch,
  describeBattleOutcome,
  isBattleFinished,
} from '../static-preview/battle-end.js';
import {
  rememberTerminalOutcome,
  snapshotForBattleEndOverlay,
} from '../static-preview/battle-end-ingest.js';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const preview = join(root, 'static-preview');

function fail(msg) {
  console.error(`FAIL: ${msg}`);
  process.exit(1);
}

function ok(msg) {
  console.log(`ok: ${msg}`);
}

const indexHtml = readFileSync(join(preview, 'index.html'), 'utf8');
if (!indexHtml.includes("base.href = '/qjp/'")) {
  fail('index.html missing runtime /qjp/ <base> bootstrap');
}

const pathsJs = readFileSync(join(preview, 'paths.js'), 'utf8');
if (!pathsJs.includes("path.startsWith('/qjp/')")) {
  fail('paths.js missing /qjp mount detection');
}

const inProgress = {
  schemaVersion: 1,
  boardSize: 19,
  endReason: 'none',
  winner: null,
  lockstepFrame: 12,
};
if (isBattleFinished(inProgress)) {
  fail('in-progress snapshot should not be finished');
}

const wipeout = {
  ...inProgress,
  endReason: 'wipeout',
  winner: 0,
  lockstepFrame: 99,
};
if (!isBattleFinished(wipeout)) {
  fail('wipeout snapshot should be finished');
}
const outcome = describeBattleOutcome(wipeout, 0);
if (outcome.headline !== '勝利' || outcome.tone !== 'win') {
  fail(`unexpected wipeout outcome: ${JSON.stringify(outcome)}`);
}

const mockPatched = applyMockVictoryPatch(inProgress, 'win', 0);
if (mockPatched.endReason !== 'wipeout' || mockPatched.winner !== 0) {
  fail('applyMockVictoryPatch(win) should set wipeout winner=0');
}

let sticky = rememberTerminalOutcome(wipeout, null);
const flickerPoll = { ...wipeout, endReason: 'none', winner: null, lockstepFrame: 100 };
const overlaySnap = snapshotForBattleEndOverlay(flickerPoll, sticky);
if (!isBattleFinished(overlaySnap)) {
  fail('sticky overlay should treat flicker poll as terminal');
}
if (overlaySnap.endReason !== 'wipeout' || overlaySnap.winner !== 0) {
  fail('sticky overlay should preserve wipeout fields');
}

ok('static-preview battle-end + /qjp base paths; wipeout/mockVictory overlay contract');
ok('Live E2E: WIPEOUT_SMOKE=1 ./scripts/janus-http-smoke.sh then /qjp/?live=1 until overlay shows 殲滅');
