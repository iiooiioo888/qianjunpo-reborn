import { LIVE_AUTO_TICK_MS } from './config.js';
import {
  deriveLocalSuggestion,
  normalizeInferSuggestion,
  summarizeBattleContext,
} from './local-suggest.js';
import { TACTICAL_COMMAND_KIND_SKILL } from './stub-skill.js';

/**
 * @typedef {{ kind: string, unitId: number, to: {x:number,y:number}, skillId?: number, tacticalKind: number, inferPath?: string }} TacticalSuggestion
 */

/**
 * @returns {Promise<{ suggestion: TacticalSuggestion | null, source: string }>}
 */
export async function fetchLiveSuggestion(boot, battleId, snap) {
  const wbUrl = `${boot.suggestWriteBackUrl}?battle_id=${encodeURIComponent(battleId)}`;
  try {
    const res = await fetch(wbUrl, { cache: 'no-store' });
    if (res.ok) {
      const json = await res.json();
      const normalized = normalizeInferSuggestion(json.suggestion, snap);
      if (normalized) {
        return { suggestion: normalized, source: 'infer-write-back' };
      }
    }
  } catch {
    /* Janus often has no /v1/suggest mirror */
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
      const normalized = normalizeInferSuggestion(json.suggestion, snap);
      if (normalized) {
        return { suggestion: normalized, source: 'infer-post' };
      }
    }
  } catch {
    /* fall through */
  }

  const local = deriveLocalSuggestion(snap);
  return { suggestion: local, source: 'local-ai' };
}

export function formatAutoHudLine(state) {
  if (!state.enabled) {
    return 'auto: off';
  }
  if (state.paused) {
    return `auto: paused (${state.pauseReason}) · next tick ~${state.tickMs}ms`;
  }
  if (state.busy) {
    return `auto: busy (${state.busyLabel})`;
  }
  const last = state.lastSource
    ? `last=${state.lastSource} ${state.lastKind ?? ''} u${state.lastUnitId ?? '?'}`
    : 'waiting';
  return `auto: on · ${last} · tick ${state.tickMs}ms`;
}

/**
 * Wires timer-driven suggest → onCommand callback (app.js submits command + lockstep).
 */
export function createLiveAutoPlayer({
  boot,
  getBattleId,
  getSnapshot,
  isBattleFinished,
  onCommand,
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

  function pauseForUser(ms = 8000, reason = 'manual') {
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
      outcome = await fetchLiveSuggestion(boot, getBattleId(), snap);
    } catch (err) {
      busy = false;
      lastState.busy = false;
      emitHud();
      onStatus?.(`auto: suggest error ${err instanceof Error ? err.message : String(err)}`);
      return;
    }

    const sug = outcome.suggestion;
    if (!sug) {
      busy = false;
      lastState.busy = false;
      emitHud();
      return;
    }

    lastState.lastSource = outcome.source;
    lastState.lastKind = sug.kind;
    lastState.lastUnitId = sug.unitId;
    lastState.busyLabel = 'command';
    emitHud();

    try {
      await onCommand(sug);
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
    window.setTimeout(() => void tick(), 400);
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
    formatLine: () => formatAutoHudLine(lastState),
  };
}

export function suggestionToCommandArgs(sug) {
  if (sug.tacticalKind === TACTICAL_COMMAND_KIND_SKILL) {
    return { kind: TACTICAL_COMMAND_KIND_SKILL, skillId: sug.skillId };
  }
  return { kind: 1 };
}
