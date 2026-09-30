import {
  CHAR_CARD_V04,
  LIVE_POLL_INTERVAL_MS,
  LOCAL_PLAYER_OWNER,
  MOCK_SNAPSHOT_URL,
  parseBootConfig,
} from './config.js';
import { TacticalBoardRenderer } from './board.js';
import {
  formatLockstepFrameLine,
  formatRateLine,
  liveLinkStateFromPollAge,
} from './hud.js';
import { applyMockMove } from './mock-move.js';
import { computeLegalMoveDestinations } from './reachability.js';

const boot = parseBootConfig();

const els = {
  rate: document.getElementById('hud-rate'),
  frame: document.getElementById('hud-frame'),
  status: document.getElementById('hud-status'),
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

function setStatus(text, isError = false) {
  els.status.textContent = text;
  els.status.classList.toggle('error', isError);
}

function refreshHud() {
  if (!snapshot) {
    return;
  }
  els.rate.textContent = formatRateLine(snapshot);
  els.frame.textContent = formatLockstepFrameLine(snapshot, syncCtx);
  els.mode.textContent = boot.live
    ? `Live: ${boot.liveUrl}`
    : `Mock: ${MOCK_SNAPSHOT_URL}  (add ?live=1 for Janus mirror)`;
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
  const res = await fetch(MOCK_SNAPSHOT_URL, { cache: 'no-store' });
  if (!res.ok) {
    throw new Error(`mock fetch HTTP ${res.status}`);
  }
  snapshot = validateSnapshot(await res.json());
  syncCtx = { source: 'mock', link: 'connected' };
  localDrift = false;
  renderer.clearSelection();
  render();
  setStatus('Mock snapshot loaded. Click your unit (blue) to preview moves.');
}

async function pollLive() {
  const started = performance.now();
  try {
    const res = await fetch(boot.liveUrl, { cache: 'no-store', mode: 'cors' });
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
    setStatus('Live Janus mirror polling (read-only; moves disabled).');
  } catch (err) {
    syncCtx = {
      source: 'live',
      link: 'error',
      pollIntervalMs: LIVE_POLL_INTERVAL_MS,
    };
    refreshHud();
    const msg = err instanceof Error ? err.message : String(err);
    setStatus(
      `Live fetch failed (${msg}). Cross-origin needs Dev nginx reverse proxy (e.g. /janus/ → Janus :18090). See README.`,
      true,
    );
  }
}

function startLive() {
  syncCtx.source = 'live';
  void pollLive();
  pollTimer = window.setInterval(() => void pollLive(), LIVE_POLL_INTERVAL_MS);
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

  if (boot.live && !localDrift) {
    setStatus('Live mode: snapshot is authoritative; local move preview disabled.');
    return;
  }

  const selected = renderer.selectedUnitId;
  if (selected != null) {
    const legal = renderer.legalCells;
    const isLegal = legal.some((c) => c.x === x && c.y === y);
    if (isLegal) {
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
    img.src = card.src;
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
      `Mock load failed (${msg}). Serve from client/ root: cd client && npx serve . — do not open file://`,
      true,
    );
  });
}

window.addEventListener('beforeunload', () => {
  if (pollTimer) {
    clearInterval(pollTimer);
  }
});
