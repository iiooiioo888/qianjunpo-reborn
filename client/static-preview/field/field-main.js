import { createBattleFromDeploy, advanceAutoTurn, tickBattleFx } from './field-battle.js';
import { drawBattlefield } from './field-render.js';
import { loadFieldAssets } from './field-assets.js';

const deployScreen = document.getElementById('screen-deploy');
const battleScreen = document.getElementById('screen-battle');
const canvas = document.getElementById('field-canvas');
const statusLine = document.getElementById('status-line');
const btnExpedition = document.getElementById('btn-expedition');
const btnBack = document.getElementById('btn-back-deploy');

/** @type {{ infantry: number, archer: number, cavalry: number }} */
let deployTroops = { infantry: 120, archer: 100, cavalry: 80 };

/** @type {ReturnType<createBattleFromDeploy> | null} */
let battleState = null;

let lastOrigin = { originX: 0, originY: 0 };
let assetsReady = false;

/** 每回合間隔（毫秒） */
const TURN_INTERVAL_MS = 750;
let turnTimer = null;
let fxFrame = null;

function readDeployFromDom() {
  document.querySelectorAll('.troop-row').forEach((row) => {
    const type = row.dataset.type;
    const input = row.querySelector('[data-input="troops"]');
    const output = row.querySelector('[data-output="troops"]');
    const v = Number(input.value);
    output.textContent = String(v);
    deployTroops[type] = v;
  });
}

document.querySelectorAll('.troop-row [data-input="troops"]').forEach((input) => {
  input.addEventListener('input', () => readDeployFromDom());
});

function stopBattleLoops() {
  if (turnTimer) {
    clearInterval(turnTimer);
    turnTimer = null;
  }
  if (fxFrame) {
    cancelAnimationFrame(fxFrame);
    fxFrame = null;
  }
}

function startFxLoop() {
  if (fxFrame) return;
  const tick = () => {
    if (!battleState) {
      fxFrame = null;
      return;
    }
    battleState = tickBattleFx(battleState);
    redraw();
    fxFrame = requestAnimationFrame(tick);
  };
  fxFrame = requestAnimationFrame(tick);
}

function startAutoTurnLoop() {
  stopBattleLoops();
  startFxLoop();
  turnTimer = setInterval(() => {
    if (!battleState || battleState.gameResult) {
      stopBattleLoops();
      syncStatus();
      redraw();
      return;
    }
    battleState = advanceAutoTurn(battleState);
    syncStatus();
    redraw();
  }, TURN_INTERVAL_MS);
}

btnExpedition.addEventListener('click', () => {
  if (!assetsReady) return;
  readDeployFromDom();
  battleState = createBattleFromDeploy(deployTroops);
  deployScreen.classList.add('hidden');
  battleScreen.classList.remove('hidden');
  syncStatus();
  redraw();
  startAutoTurnLoop();
});

btnBack.addEventListener('click', () => {
  stopBattleLoops();
  battleState = null;
  battleScreen.classList.add('hidden');
  deployScreen.classList.remove('hidden');
});

function syncStatus() {
  if (statusLine && battleState) {
    statusLine.textContent = battleState.message;
  }
}

function redraw() {
  if (!battleState || !assetsReady) return;
  const ctx = canvas.getContext('2d');
  lastOrigin = drawBattlefield(ctx, canvas, battleState);
}

window.addEventListener('resize', () => redraw());

loadFieldAssets()
  .then(() => {
    assetsReady = true;
    btnExpedition.disabled = false;
    readDeployFromDom();
  })
  .catch((err) => {
    console.error(err);
    if (statusLine) {
      statusLine.textContent = '手繪資源載入失敗，請重新整理。';
    }
  });

readDeployFromDom();
btnExpedition.disabled = true;
