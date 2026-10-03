import { LIVE_AUTO_TICK_MS } from './config.js';
import {
  deriveLocalCommand,
  normalizeTacticalCommand,
  summarizeBattleContext,
} from './local-suggest.js';

/**
 * @returns {Promise<{ command: object | null, source: string }>}
 */
export async function fetchLiveCommand(boot, battleId, sessionId, snap) {
  const wbUrl = `${boot.suggestWriteBackUrl}?battle_id=${encodeURIComponent(battleId)}`;
  try {
    const res = await fetch(wbUrl, { cache: 'no-store' });
    if (res.ok) {
      const json = await res.json();
      const cmd = normalizeTacticalCommand(json.command, battleId, sessionId, snap);
      if (cmd) {
        return { command: cmd, source: 'infer-write-back' };
      }
    }
  } catch {
    /* suggest mirror may be absent on Janus host */
  }

  const ctx = summarizeBattleContext(snap);
  try {
    const res = await fetch(boot.suggestUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        battle_id: battleId,
        persona: 'guard',
        order: 'advance and strike when in range',
        battle_context: ctx,
      }),
    });
    if (res.ok) {
      const json = await res.json();
      const cmd = normalizeTacticalCommand(json.command, battleId, sessionId, snap);
      if (cmd) {
        return { command: cmd, source: 'infer-post' };
      }
    }
  } catch {
    /* fall through */
  }

  const local = deriveLocalCommand(battleId, sessionId, snap);
  return { command: local, source: 'local-ai' };
}

export function formatAutoHudLine(state) {
  if (!state.enabled) {
    return 'auto: off';
  }
  if (state.paused) {
    return `auto: paused (${state.pauseReason}) · tick ${state.tickMs}ms`;
  }
  if (state.busy) {
    return `auto: busy (${state.busyLabel})`;
  }
  const last = state.lastSource
    ? `last=${state.lastSource} kind=${state.lastKind ?? '?'} u${state.lastUnitId ?? '?'}`
    : 'waiting';
  return `auto: on · ${last} · tick ${state.tickMs}ms`;
}

export function createLiveAutoPlayer({
  boot,
  getBattleId,
  getSessionId,
  getSnapshot,
  isBattleFinished,
  onCommand,
  onStep,
  onStatus,
  tickMs = LIVE_AUTO_TICK_MS,
}) {
  let timer = null;
  let busy = false;
  let pausedUntil = 0;
  let pauseReason = '';
  let lastState = {
    enabled: Boolean(boot.liveAuto),
    paused: false,
    pauseReason: '',
    busy: false,
    busyLabel: '',
    tickMs,
    lastSource: '',
    lastKind: '',
    lastUnitId: null,
  };

  function emitHud() {
    onStatus?.(formatAutoHudLine(lastState));
  }

  function pauseForUser(ms = 8000, reason = 'player') {
    pausedUntil = Date.now() + ms;
    pauseReason = reason;
    lastState.paused = true;
    lastState.pauseReason = reason;
    emitHud();
  }

  async function tick() {
    if (!boot.liveAuto || busy) {
      return;
    }
    const snap = getSnapshot();
    if (!snap || isBattleFinished(snap)) {
      stop();
      return;
    }
    if (Date.now() < pausedUntil) {
      lastState.paused = true;
      emitHud();
      return;
    }
    lastState.paused = false;
    lastState.pauseReason = '';

    busy = true;
    lastState.busy = true;
    lastState.busyLabel = 'suggest';
    emitHud();

    let outcome;
    try {
      outcome = await fetchLiveCommand(
        boot,
        getBattleId(),
        getSessionId(),
        snap,
      );
    } catch (err) {
      busy = false;
      lastState.busy = false;
      emitHud();
      onStatus?.(`auto: suggest error ${err instanceof Error ? err.message : String(err)}`);
      return;
    }

    const cmd = outcome.command;
    if (cmd) {
      lastState.lastSource = outcome.source;
      lastState.lastKind = String(cmd.kind);
      lastState.lastUnitId = cmd.unit_id;
      lastState.busyLabel = 'command';
      emitHud();
      try {
        await onCommand(cmd);
      } catch {
        /* onCommand surfaces HUD errors */
      }
    }

    lastState.busyLabel = 'step';
    emitHud();
    try {
      await onStep();
    } finally {
      busy = false;
      lastState.busy = false;
      emitHud();
    }
  }

  function start() {
    if (!boot.liveAuto || timer != null) {
      return;
    }
    lastState.enabled = true;
    emitHud();
    timer = window.setInterval(() => void tick(), tickMs);
    window.setTimeout(() => void tick(), 500);
  }

  function stop() {
    if (timer) {
      clearInterval(timer);
      timer = null;
    }
    lastState.enabled = false;
    emitHud();
  }

  return {
    start,
    stop,
    pauseForUser,
    notifyUserAction() {
      pauseForUser(8000, 'player');
    },
    isBusy: () => busy,
  };
}
