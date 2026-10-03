import { createBattleFromDeploy, handleCellClick } from './field-battle.js';
import { drawBattlefield } from './field-render.js';
import { pickGridCell } from './field-iso.js';

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

btnExpedition.addEventListener('click', () => {
  readDeployFromDom();
  battleState = createBattleFromDeploy(deployTroops);
  deployScreen.classList.add('hidden');
  battleScreen.classList.remove('hidden');
  syncStatus();
  redraw();
});

btnBack.addEventListener('click', () => {
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
  if (!battleState) return;
  const ctx = canvas.getContext('2d');
  lastOrigin = drawBattlefield(ctx, canvas, battleState);
}

canvas.addEventListener('click', (ev) => {
  if (!battleState) return;
  const rect = canvas.getBoundingClientRect();
  const scaleX = canvas.width / rect.width;
  const scaleY = canvas.height / rect.height;
  const px = (ev.clientX - rect.left) * scaleX;
  const py = (ev.clientY - rect.top) * scaleY;
  const cell = pickGridCell(px, py, lastOrigin.originX, lastOrigin.originY);
  if (!cell) {
    battleState = { ...battleState, message: '請點在格子上。' };
    syncStatus();
    return;
  }

  battleState = handleCellClick(battleState, cell.x, cell.y);
  syncStatus();
  redraw();
});

window.addEventListener('resize', () => redraw());

readDeployFromDom();
