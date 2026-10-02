import { isBattleFinished } from './battle-end.js';

/**
 * Remember the latest terminal outcome for overlay display (Live wipeout polls).
 * @param {object | null | undefined} snap
 * @param {object | null} stickyTerminal
 */
export function rememberTerminalOutcome(snap, stickyTerminal) {
  if (!snap || !isBattleFinished(snap)) {
    return stickyTerminal;
  }
  return {
    winner: snap.winner,
    endReason: snap.endReason,
    lockstepFrame: snap.lockstepFrame,
  };
}

/**
 * Overlay stays visible when a transient poll omits terminal fields on the live snapshot.
 * @param {object | null | undefined} snap
 * @param {object | null} stickyTerminal
 */
export function snapshotForBattleEndOverlay(snap, stickyTerminal) {
  if (!snap) {
    return null;
  }
  if (isBattleFinished(snap)) {
    return snap;
  }
  if (!stickyTerminal) {
    return snap;
  }
  return {
    ...snap,
    winner: stickyTerminal.winner,
    endReason: stickyTerminal.endReason,
    lockstepFrame: stickyTerminal.lockstepFrame ?? snap.lockstepFrame,
  };
}
