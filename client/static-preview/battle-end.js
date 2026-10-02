/** Sentinel aligned with `tactical.NoWinner` (255). */
export const NO_WINNER = 255;

const END_REASON_ZH = {
  wipeout: '殲滅（全滅）',
  timeout: '回合超時',
  mutual_wipe: '雙方同滅',
  occupy: '占點勝利',
  none: '進行中',
};

/**
 * Battle ended when authoritative snapshot reports a terminal endReason.
 * In-progress snapshots use `endReason: "none"` and `winner: null`.
 * @param {object | null | undefined} snap
 */
export function isBattleFinished(snap) {
  if (!snap) {
    return false;
  }
  const reason = snap.endReason;
  return typeof reason === 'string' && reason.length > 0 && reason !== 'none';
}

/**
 * @param {string} endReason
 */
export function formatEndReasonZh(endReason) {
  if (!endReason || endReason === 'none') {
    return '—';
  }
  return END_REASON_ZH[endReason] ?? endReason;
}

/**
 * @param {object} snap finished snapshot
 * @param {number} localPlayerOwner e.g. 0
 * @returns {{ headline: string, detail: string, tone: 'win' | 'lose' | 'draw' }}
 */
export function describeBattleOutcome(snap, localPlayerOwner) {
  const reasonZh = formatEndReasonZh(snap.endReason);
  const w = snap.winner;
  if (w === NO_WINNER) {
    return {
      headline: '和局',
      detail: reasonZh,
      tone: 'draw',
    };
  }
  if (typeof w === 'number') {
    if (w === localPlayerOwner) {
      return {
        headline: '勝利',
        detail: reasonZh,
        tone: 'win',
      };
    }
    return {
      headline: '敗北',
      detail: reasonZh,
      tone: 'lose',
    };
  }
  return {
    headline: '戰局結束',
    detail: reasonZh,
    tone: 'draw',
  };
}

/**
 * Dev/mock: patch snapshot with terminal winner/endReason (`?mockVictory=win|lose|draw`).
 * @param {object} snap
 * @param {'win' | 'lose' | 'draw' | ''} mode
 * @param {number} localPlayerOwner
 */
export function applyMockVictoryPatch(snap, mode, localPlayerOwner) {
  if (!mode || !snap) {
    return snap;
  }
  const patched = { ...snap };
  if (mode === 'win') {
    patched.winner = localPlayerOwner;
    patched.endReason = 'wipeout';
  } else if (mode === 'lose') {
    patched.winner = localPlayerOwner === 0 ? 1 : 0;
    patched.endReason = 'wipeout';
  } else if (mode === 'draw') {
    patched.winner = NO_WINNER;
    patched.endReason = 'mutual_wipe';
  }
  return patched;
}
