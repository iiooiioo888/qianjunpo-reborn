import {
  CHAR_CARD_V04,
  LIVE_POLL_INTERVAL_MS,
  LOCAL_PLAYER_OWNER,
  mockSnapshotUrl,
  parseBootConfig,
} from './config.js';
import { resolveAppUrl } from './paths.js';
import { TacticalBoardRenderer } from './board.js';
import {
  formatLockstepFrameLine,
  formatRateLine,
  liveLinkStateFromPollAge,
} from './hud.js';
import { applyMockMove } from './mock-move.js';
import { computeLegalMoveDestinations } from './reachability.js';
import { prepareLiveJanusSession } from './live-gateway.js';

const boot = parseBootConfig();

const els = {
  rate: document.getElementById('hud-rate'),
  frame: document.getElementById('hud-frame'),
  status: document.getElementById('hud-status'),
  retryLive: document.getElementById('hud-retry-live'),
  mode: document.getElementById('hud-mode'),
  cards: document.getElementById('char-cards'),
  canvas: document.getElementById('board'),
};

const renderer = new TacticalBoardRenderer(els.canvas);

/** @type {import('./types.js').ViewSnapshot | null} */
let snapshot = null;
let syncCtx = {
  source: boot.live ? 'live' : 'mock',
  link: 'connected',
  pollAgeMs: undefined,
  pollIntervalMs: LIVE_POLL_INTERVAL_MS,
};
let lastPollAt = 0;
let pollTimer = null;
let localDrift = false;
let liveSessionId = '';
let liveBattleId = battleIdFromLiveUrl(boot.liveUrlRaw);

function setStatus(text, isError = false) {
  els.status.textContent = text;
  els.status.classList.toggle('error', isError);
}

function hideLivePrepareRetry() {
  els.retryLive.hidden = true;
  els.retryLive.onclick = null;
}

function showLivePrepareRetry(onRetry) {
  els.retryLive.hidden = false;
  els.retryLive.onclick = () => {
    hideLivePrepareRetry();
    onRetry();
  };
}

function stopLivePoll() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

function refreshHud() {
  if (!snapshot) {
    return;
  }
  els.rate.textContent = formatRateLine(snapshot);
  els.frame.textContent = formatLockstepFrameLine(snapshot, syncCtx);
  els.mode.textContent = boot.live
    ? `Live: ${boot.liveUrlRaw}`
    : `Mock: mock/demo_initial.json  ·  ?live=1 → ${boot.liveUrlRaw}`;
}

function render() {
  if (snapshot) {
    renderer.draw(snapshot);
    refreshHud();
  }
}

function validateSnapshot(raw) {
  if (!raw || raw.schemaVersion !== 1 || raw.boardSize !== 19) {
    throw new Error('unsupported view snapshot schema');
  }
  if (!Array.isArray(raw.cells) || raw.cells.length !== 19) {
    throw new Error('cells grid size mismatch');
  }
  return raw;
}

async function loadMock() {
  const url = mockSnapshotUrl();
  const res = await fetch(url, { cache: 'no-store' });
  if (!res.ok) {
    throw new Error(`${url} HTTP ${res.status}`);
  }
  snapshot = validateSnapshot(await res.json());
  syncCtx = { source: 'mock', link: 'connected' };
  localDrift = false;
  renderer.clearSelection();
  render();
  setStatus(`Mock snapshot loaded (${url}). Click your unit (blue) to preview moves.`);
}

function battleIdFromLiveUrl(liveUrlRaw) {
  try {
    const q = liveUrlRaw.includes('?') ? liveUrlRaw.slice(liveUrlRaw.indexOf('?')) : '';
    const id = new URLSearchParams(q).get('battle_id');
    return id || 'default/0';
  } catch {
    return 'default/0';
  }
}

function liveSnapshotUrlForBattle(battleId) {
  const raw = boot.liveUrlRaw;
  if (raw.includes('battle_id=')) {
    return resolveAppUrl(raw.replace(/battle_id=[^&]+/, `battle_id=${encodeURIComponent(battleId)}`));
  }
  const sep = raw.includes('?') ? '&' : '?';
  return resolveAppUrl(`${raw}${sep}battle_id=${encodeURIComponent(battleId)}`);
}

async function submitLiveCommand(unitId, to) {
  const unit = snapshot?.units.find((u) => u.id === unitId && u.hp > 0);
  if (!unit) {
    return { accepted: false, rejectReason: 'unit missing' };
  }
  const body = {
    battle_id: liveBattleId,
    player_id: unit.owner,
    kind: 1,
    unit_id: unitId,
    to_x: to.x,
    to_y: to.y,
    session_id: liveSessionId,
  };
  const res = await fetch(boot.commandUrl, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const text = await res.text();
    return { accepted: false, rejectReason: `HTTP ${res.status}: ${text}` };
  }
  const json = await res.json();
  return {
    accepted: Boolean(json.accepted),
    rejectReason: json.reject_reason,
    lockstepFrame: json.lockstep_frame,
    stateHash: json.state_hash,
  };
}

async function pollLive() {
  const started = performance.now();
  try {
    const res = await fetch(liveSnapshotUrlForBattle(liveBattleId), {
      cache: 'no-store',
      mode: 'cors',
    });
    if (!res.ok) {
      throw new Error(`live HTTP ${res.status}`);
    }
    const body = await res.json();
    snapshot = validateSnapshot(body);
    lastPollAt = Date.now();
    const pollAgeMs = performance.now() - started;
    syncCtx = {
      source: 'live',
      link: localDrift ? 'stale' : liveLinkStateFromPollAge(pollAgeMs, LIVE_POLL_INTERVAL_MS),
      pollAgeMs,
      pollIntervalMs: LIVE_POLL_INTERVAL_MS,
    };
    if (!localDrift) {
      renderer.clearSelection();
    }
    render();
    setStatus('Live Janus mirror polling; legal moves POST to v1/tactical/command.');
  } catch (err) {
    syncCtx = {
      source: 'live',
      link: 'error',
      pollIntervalMs: LIVE_POLL_INTERVAL_MS,
    };
    refreshHud();
    const msg = err instanceof Error ? err.message : String(err);
    setStatus(
      `Live fetch failed (${msg}). Use same-origin v1/… (see README: /qjp/v1/ or :18093 /v1/).`,
      true,
    );
  }
}

async function runLivePrepare() {
  stopLivePoll();
  lastPollAt = 0;
  hideLivePrepareRetry();
  setStatus('Live：POST connect → enter-battle…');
  try {
    const prepared = await prepareLiveJanusSession(boot.liveGateway, liveBattleId);
    liveSessionId = prepared.sessionId;
    liveBattleId = prepared.battleId;
    if (prepared.initialSnapshot) {
      snapshot = validateSnapshot(prepared.initialSnapshot);
      render();
    }
    setStatus(`Live：EnterBattle 已建局 ${liveBattleId}；開始快照輪詢。`);
    await pollLive();
    pollTimer = window.setInterval(() => void pollLive(), LIVE_POLL_INTERVAL_MS);
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    syncCtx = { source: 'live', link: 'error', pollIntervalMs: LIVE_POLL_INTERVAL_MS };
    refreshHud();
    setStatus(msg, true);
    showLivePrepareRetry(() => {
      void runLivePrepare();
    });
  }
}

function startLive() {
  syncCtx.source = 'live';
  void runLivePrepare();
}

function unitAt(x, y) {
  if (!snapshot) {
    return null;
  }
  const cell = snapshot.cells[y]?.[x];
  if (!cell || cell.unitId === 0) {
    return null;
  }
  return snapshot.units.find((u) => u.id === cell.unitId && u.hp > 0) ?? null;
}

function onCellClick(x, y) {
  if (!snapshot) {
    return;
  }
  const unitAtCell = unitAt(x, y);

  const selected = renderer.selectedUnitId;
  if (selected != null) {
    const legal = renderer.legalCells;
    const isLegal = legal.some((c) => c.x === x && c.y === y);
    if (isLegal) {
      if (boot.live && !localDrift) {
        void (async () => {
          setStatus('Live：送出移動指令…');
          try {
            const result = await submitLiveCommand(selected, { x, y });
            if (result.accepted) {
              const hashPart =
                result.stateHash != null ? ` hash=${result.stateHash}` : '';
              renderer.clearSelection();
              setStatus(
                `Live：指令已接受 frame=${result.lockstepFrame ?? '?'}${hashPart}（等待快照）`,
              );
              await pollLive();
            } else {
              setStatus(`Live 拒絕：${result.rejectReason ?? 'unknown'}`, true);
            }
          } catch (err) {
            const msg = err instanceof Error ? err.message : String(err);
            setStatus(`Live 指令失敗：${msg}`, true);
          }
        })();
        return;
      }
      const result = applyMockMove(snapshot, selected, { x, y });
      if (result.ok && result.snapshot) {
        snapshot = result.snapshot;
        localDrift = boot.live;
        syncCtx = boot.live
          ? { source: 'live', link: 'stale', pollIntervalMs: LIVE_POLL_INTERVAL_MS }
          : { source: 'mock', link: 'stale' };
        renderer.clearSelection();
        render();
        setStatus(
          boot.live
            ? 'Local mock move applied on top of live (non-authoritative drift).'
            : 'Mock move applied (non-authoritative).',
        );
      }
      return;
    }
  }

  if (unitAtCell && unitAtCell.owner === LOCAL_PLAYER_OWNER) {
    if (selected === unitAtCell.id) {
      renderer.clearSelection();
      render();
      return;
    }
    const legalCells = computeLegalMoveDestinations(snapshot, unitAtCell.id);
    renderer.setSelection(unitAtCell.id, legalCells);
    render();
    setStatus(`Unit ${unitAtCell.id} selected — ${legalCells.length} legal cells (move≤4).`);
    return;
  }

  renderer.clearSelection();
  render();
}

function bindCards() {
  if (!boot.showCards) {
    els.cards.hidden = true;
    return;
  }
  els.cards.innerHTML = '';
  for (const card of CHAR_CARD_V04) {
    const fig = document.createElement('figure');
    fig.className = 'char-card';
    const img = document.createElement('img');
    img.src = resolveAppUrl(card.src);
    img.alt = card.label;
    img.loading = 'lazy';
    img.onerror = () => {
      img.replaceWith(Object.assign(document.createElement('div'), {
        className: 'char-placeholder',
        textContent: card.label,
      }));
    };
    const cap = document.createElement('figcaption');
    cap.textContent = card.label;
    fig.append(img, cap);
    els.cards.append(fig);
  }
}

els.canvas.addEventListener('click', (ev) => {
  const rect = els.canvas.getBoundingClientRect();
  const scaleX = els.canvas.width / rect.width;
  const scaleY = els.canvas.height / rect.height;
  const px = (ev.clientX - rect.left) * scaleX;
  const py = (ev.clientY - rect.top) * scaleY;
  const cell = renderer.cellAtPixel(px, py);
  if (cell) {
    onCellClick(cell.x, cell.y);
  }
});

bindCards();

if (boot.live) {
  startLive();
} else {
  loadMock().catch((err) => {
    syncCtx = { source: 'mock', link: 'error' };
    const msg = err instanceof Error ? err.message : String(err);
    setStatus(
      `Mock load failed (${msg}). Serve this folder: cd client/static-preview && npx serve . — or use HTTP, not file://`,
      true,
    );
  });
}

window.addEventListener('beforeunload', () => {
  if (pollTimer) {
    clearInterval(pollTimer);
  }
});
