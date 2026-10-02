import { describeBattleOutcome, isBattleFinished } from './battle-end.js';

/**
 * @param {{ overlay: HTMLElement, headline: HTMLElement, detail: HTMLElement, meta: HTMLElement }} els
 */
export function createBattleEndOverlayElements(root) {
  const overlay = document.createElement('div');
  overlay.id = 'battle-end-overlay';
  overlay.className = 'battle-end-overlay';
  overlay.hidden = true;
  overlay.setAttribute('role', 'dialog');
  overlay.setAttribute('aria-modal', 'true');
  overlay.setAttribute('aria-label', '戰局結束');

  const panel = document.createElement('div');
  panel.className = 'battle-end-panel';

  const headline = document.createElement('p');
  headline.className = 'battle-end-headline';
  headline.id = 'battle-end-headline';

  const detail = document.createElement('p');
  detail.className = 'battle-end-detail';
  detail.id = 'battle-end-detail';

  const meta = document.createElement('p');
  meta.className = 'battle-end-meta';
  meta.id = 'battle-end-meta';

  panel.append(headline, detail, meta);
  overlay.append(panel);
  root.append(overlay);

  return { overlay, headline, detail, meta };
}

/**
 * @param {object | null} snap
 * @param {{ overlay: HTMLElement, headline: HTMLElement, detail: HTMLElement, meta: HTMLElement, localPlayerOwner: number }} ctx
 */
export function updateBattleEndOverlay(snap, ctx) {
  const { overlay, headline, detail, meta, localPlayerOwner } = ctx;
  if (!isBattleFinished(snap)) {
    overlay.hidden = true;
    overlay.classList.remove('tone-win', 'tone-lose', 'tone-draw');
    return;
  }
  const outcome = describeBattleOutcome(snap, localPlayerOwner);
  headline.textContent = outcome.headline;
  detail.textContent = `結束原因：${outcome.detail}`;
  const winnerLabel =
    snap.winner === 255 ? '和局 (255)' : typeof snap.winner === 'number' ? `玩家 ${snap.winner}` : '—';
  meta.textContent = `winner=${winnerLabel} · endReason=${snap.endReason ?? '—'} · frame ${snap.lockstepFrame ?? '?'}`;
  overlay.classList.remove('tone-win', 'tone-lose', 'tone-draw');
  overlay.classList.add(
    outcome.tone === 'win' ? 'tone-win' : outcome.tone === 'lose' ? 'tone-lose' : 'tone-draw',
  );
  overlay.hidden = false;
}
