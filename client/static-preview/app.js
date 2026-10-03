import { captureBootQuery } from './boot-query.js';
import {
  CHAR_CARD_V04,
  LIVE_POLL_INTERVAL_MS,
  LOCAL_PLAYER_OWNER,
  mockSnapshotUrl,
  parseBootConfig,
} from './config.js';
import { createCampaignFlow } from './campaign-flow.js';
import { resolveAppUrl } from './paths.js';
import { TacticalBoardRenderer } from './board.js';
import { applyMockVictoryPatch, isBattleFinished } from './battle-end.js';
import {
  rememberTerminalOutcome,
  snapshotForBattleEndOverlay,
} from './battle-end-ingest.js';
import {
  createBattleEndOverlayElements,
  updateBattleEndOverlay,
} from './battle-end-overlay.js';
import {
  formatLastSkillCastLine,
  formatLockstepFrameLine,
  formatRateLine,
  formatSelectionLine,
  lastSkillCastHudClass,
  liveLinkStateFromPollAge,
  pickLastSkillCast,
} from './hud.js';
import { applyMockMove } from './mock-move.js';
import { computeLegalMoveDestinations } from './reachability.js';
import {
  findStubStrikeTarget,
  isInAttackRange,
  TACTICAL_COMMAND_KIND_SKILL,
} from './stub-skill.js';
import {
  hasLiveAccessToken,
  LIVE_ACCESS_TOKEN_MISSING_MESSAGE,
  mintLiveAccessToken,
  prepareLiveJanusSession,
  stepTacticalLockstep,
  TACTICAL_AUTO_COMMAND_MODE_BOTH,
  TACTICAL_LOCKSTEP_STEPS_AFTER_COMMAND,
} from './live-gateway.js';
import { createLiveAutoPlayer } from './live-auto-play.js';
import {
  ensureFreshLiveZoneBeforeEnter,
  isTerminalLiveSnapshot,
  mintFreshLiveZoneId,
  persistZoneIdInLocation,
} from './live-zone-entry.js';
import {
  CAMPAIGN_BUTTON_ICONS,
  HUD_RESOURCE_ICONS,
  HUD_TOWN_ICONS,
} from './asset-registry.js';

captureBootQuery();
const boot = parseBootConfig();

const els = {
  rate: document.getElementById('hud-rate'),
  frame: document.getElementById('hud-frame'),
  selection: document.getElementById('hud-selection'),
  status: document.getElementById('hud-status'),
  retryLive: document.getElementById('hud-retry-live'),
  mintToken: document.getElementById('hud-mint-token'),
  castSkill: document.getElementById('hud-cast-skill'),
  skillCast: document.getElementById('hud-skill-cast'),
  mode: document.getElementById('hud-mode'),
  auto: document.getElementById('hud-auto'),
  resIcons: document.getElementById('hud-res-icons'),
  townIcons: document.getElementById('hud-town-icons'),
  campaignPhase: document.getElementById('campaign-phase'),
  campaignDeploy: document.getElementById('campaign-deploy'),
  campaignSupply: document.getElementById('campaign-supply'),
  campaignBattle: document.getElementById('campaign-battle'),
  btnCampaignSupply: document.getElementById('btn-campaign-supply'),
  btnCampaignBattle: document.getElementById('btn-campaign-battle'),
  cards: document.getElementById('char-cards'),
  canvas: document.getElementById('board'),
  boardWrap: document.getElementById('board-wrap'),
};

const battleEndOverlay = createBattleEndOverlayElements(els.boardWrap);

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
/** @type {object | null} sticky lastSkillCast when polls omit the field briefly */
let stickyLastSkillCast = null;
/** @type {object | null} sticky terminal winner/endReason for overlay (Live wipeout polls) */
let stickyTerminalOutcome = null;

let campaign = null;

const liveAutoPlayer = boot.live
  ? createLiveAutoPlayer({
      boot,
      getBattleId: () => liveBattleId,
      getSessionId: () => liveSessionId,
      getSnapshot: () => snapshot,
      isBattleFinished,
      onStatus: (line) => {
        if (els.auto) {
          els.auto.textContent = line;
        }
      },
      onCommand: async (cmdBody) => {
        const result = await submitLiveTacticalCommandBody(cmdBody);
        if (!result.accepted) {
          setStatus(formatLiveCommandError(result), true);
        }
      },
      onStep: async () => {
        await applyLiveLockstepTick(1);
      },
    })
  : null;

function setStatus(text, isError = false) {
  els.status.textContent = text;
  els.status.classList.toggle('error', isError);
}

function hideMintTokenButton() {
  els.mintToken.hidden = true;
}

function showMintTokenButton() {
  els.mintToken.hidden = false;
  els.mintToken.textContent = LIVE_LABEL_MINT_TOKEN;
}

function hideLiveRetry() {
  liveRetryKind = 'none';
  els.retryLive.hidden = true;
  els.retryLive.onclick = null;
  els.retryLive.textContent = '';
}

function showLiveRetry(kind, label, onRetry) {
  liveRetryKind = kind;
  els.retryLive.hidden = false;
  els.retryLive.textContent = label;
  els.retryLive.onclick = () => {
    if (kind !== 'command') {
      hideLiveRetry();
    }
    onRetry();
  };
}

function clearCommandFailure() {
  pendingCommandFailure = null;
  if (liveRetryKind === 'command') {
    hideLiveRetry();
  }
}

function restoreCommandFailureHud() {
  if (!pendingCommandFailure) {
    return;
  }
  setStatus(pendingCommandFailure.message, true);
  showLiveRetry('command', LIVE_LABEL_COMMAND_RETRY, () => {
    void runFailedCommandRetry();
  });
}

const LIVE_LABEL_MINT_TOKEN = '一鍵取 token';
const LIVE_LABEL_PREPARE_RETRY = '重試 Live 建局（connect → enter-battle）';
const LIVE_LABEL_POLL_RECONNECT = '重連 Live';
const LIVE_LABEL_COMMAND_RETRY = '重試戰術指令';

/** @type {'none' | 'prepare' | 'poll' | 'command'} */
let liveRetryKind = 'none';
/** @type {null | { unitId: number, to: { x: number, y: number }, kind: number, skillId?: number, message: string }} */
let pendingCommandFailure = null;

function stopLivePoll() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
  liveAutoPlayer?.stop();
}

function ingestSnapshot(raw) {
  let snap = validateSnapshot(raw);
  if (boot.mockVictory && !boot.live) {
    snap = applyMockVictoryPatch(snap, boot.mockVictory, LOCAL_PLAYER_OWNER);
  }
  stickyTerminalOutcome = rememberTerminalOutcome(snap, stickyTerminalOutcome);
  if (snap.lastSkillCast) {
    stickyLastSkillCast = snap.lastSkillCast;
  }
  snapshot = snap;
  if (boot.live && isBattleFinished(snapshot)) {
    stopLivePoll();
    liveAutoPlayer?.stop();
  }
}

function refreshHud() {
  if (!snapshot) {
    return;
  }
  els.rate.textContent = formatRateLine(snapshot);
  els.frame.textContent = formatLockstepFrameLine(snapshot, syncCtx);
  els.selection.textContent = formatSelectionLine(snapshot, renderer.selectedUnitId);
  els.skillCast.textContent = formatLastSkillCastLine(snapshot, stickyLastSkillCast);
  els.skillCast.className = lastSkillCastHudClass(snapshot, stickyLastSkillCast);
  updateBattleEndOverlay(snapshotForBattleEndOverlay(snapshot, stickyTerminalOutcome), {
    ...battleEndOverlay,
    localPlayerOwner: LOCAL_PLAYER_OWNER,
  });
  updateCastSkillButton();
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
  ingestSnapshot(await res.json());
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

function updateCastSkillButton() {
  if (!boot.live || !boot.debugMoves) {
    els.castSkill.hidden = true;
    return;
  }
  const selectedId = renderer.selectedUnitId;
  if (isBattleFinished(snapshot)) {
    els.castSkill.hidden = true;
    return;
  }
  if (selectedId == null || !snapshot) {
    els.castSkill.hidden = true;
    return;
  }
  const caster = snapshot.units.find((u) => u.id === selectedId && u.hp > 0);
  if (!caster || caster.owner !== LOCAL_PLAYER_OWNER) {
    els.castSkill.hidden = true;
    return;
  }
  els.castSkill.hidden = false;
  const target = findStubStrikeTarget(snapshot, caster);
  els.castSkill.disabled = target == null;
  els.castSkill.title = target
    ? `POST kind=${TACTICAL_COMMAND_KIND_SKILL} skill_id=${boot.stubSkillId} → (${target.x},${target.y})`
    : '射程內無敵方單位（可先移動靠近）';
}

async function submitLiveTacticalCommandBody(body) {
  const payload = {
    ...body,
    battle_id: body.battle_id ?? liveBattleId,
    session_id: body.session_id ?? liveSessionId,
  };
  try {
    const res = await fetch(boot.commandUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
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
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    return { accepted: false, rejectReason: `網路錯誤：${msg}` };
  }
}

async function submitLiveTacticalCommand(unitId, to, { kind = 1, skillId } = {}) {
  const unit = snapshot?.units.find((u) => u.id === unitId && u.hp > 0);
  if (!unit) {
    return { accepted: false, rejectReason: 'unit missing' };
  }
  const body = {
    battle_id: liveBattleId,
    player_id: unit.owner,
    kind,
    unit_id: unitId,
    to_x: to.x,
    to_y: to.y,
    session_id: liveSessionId,
  };
  if (kind === TACTICAL_COMMAND_KIND_SKILL) {
    body.skill_id = skillId ?? boot.stubSkillId;
  }
  return submitLiveTacticalCommandBody(body);
}

function formatLiveCommandError(result) {
  if (result.rejectReason) {
    return `Live 指令失敗：${result.rejectReason}`;
  }
  return 'Live 指令失敗：unknown';
}

function showCommandFailure(unitId, to, message, { kind = 1, skillId } = {}) {
  pendingCommandFailure = { unitId, to, kind, skillId, message };
  setStatus(message, true);
  showLiveRetry('command', LIVE_LABEL_COMMAND_RETRY, () => {
    void runFailedCommandRetry();
  });
}

async function runFailedCommandRetry() {
  if (!pendingCommandFailure) {
    return;
  }
  const { unitId, to, kind, skillId } = pendingCommandFailure;
  const label = kind === TACTICAL_COMMAND_KIND_SKILL ? '技能' : '移動';
  setStatus(`Live：送出${label}指令…`);
  const result = await submitLiveTacticalCommand(unitId, to, { kind, skillId });
  if (result.accepted) {
    await handleAcceptedLiveCommand(result);
    return;
  }
  showCommandFailure(unitId, to, formatLiveCommandError(result), { kind, skillId });
}

function formatLiveStepError(step) {
  if (step.text) {
    return `Live step-lockstep 失敗：HTTP ${step.status} ${step.text}`;
  }
  return `Live step-lockstep 失敗：HTTP ${step.status}`;
}

async function applyLiveLockstepTick(steps = 1, { autoCommandMode } = {}) {
  const step = await stepTacticalLockstep(boot.stepLockstepUrl, {
    sessionId: liveSessionId,
    battleId: liveBattleId,
    steps,
    autoCommandMode,
  });
  if (!step.ok) {
    throw new Error(formatLiveStepError(step));
  }
  if (step.viewSnapshot) {
    ingestSnapshot(step.viewSnapshot);
    lastPollAt = Date.now();
    syncCtx = {
      source: 'live',
      link: localDrift ? 'stale' : 'connected',
      pollIntervalMs: LIVE_POLL_INTERVAL_MS,
    };
    renderer.clearSelection();
    render();
  }
  return step;
}

async function enableLiveAutoCommandModeBoth() {
  await applyLiveLockstepTick(1, { autoCommandMode: TACTICAL_AUTO_COMMAND_MODE_BOTH });
}

async function applyLiveLockstepAfterCommand() {
  setStatus('Live：POST step-lockstep（步進權威帧）…');
  const step = await applyLiveLockstepTick(TACTICAL_LOCKSTEP_STEPS_AFTER_COMMAND);
  const hashPart = step.stateHash != null ? ` hash=${step.stateHash}` : '';
  setStatus(
    `Live：lockstep 已步進 frame=${step.lockstepFrame ?? '?'}${hashPart}（恢復快照輪詢）`,
  );
  await pollLive();
}

async function handleAcceptedLiveCommand(result) {
  clearCommandFailure();
  const hashPart = result.stateHash != null ? ` hash=${result.stateHash}` : '';
  renderer.clearSelection();
  setStatus(`Live：指令已接受 frame=${result.lockstepFrame ?? '?'}${hashPart}（步進 lockstep）`);
  try {
    await applyLiveLockstepAfterCommand();
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    syncCtx = {
      source: 'live',
      link: 'error',
      pollIntervalMs: LIVE_POLL_INTERVAL_MS,
    };
    refreshHud();
    setStatus(`${msg}。可點「重連 Live」或「重試戰術指令」。`, true);
    showLiveRetry('poll', LIVE_LABEL_POLL_RECONNECT, () => {
      void runLivePrepare();
    });
  }
}

async function runLiveMoveCommand(unitId, to) {
  setStatus('Live：送出移動指令…');
  const result = await submitLiveTacticalCommand(unitId, to, { kind: 1 });
  if (result.accepted) {
    await handleAcceptedLiveCommand(result);
    return;
  }
  showCommandFailure(unitId, to, formatLiveCommandError(result), { kind: 1 });
}

async function runLiveSkillCommand(unitId, to, skillId = boot.stubSkillId) {
  setStatus(`Live：送出技能指令（kind=${TACTICAL_COMMAND_KIND_SKILL} skill_id=${skillId}）…`);
  const result = await submitLiveTacticalCommand(unitId, to, {
    kind: TACTICAL_COMMAND_KIND_SKILL,
    skillId,
  });
  if (result.accepted) {
    await handleAcceptedLiveCommand(result);
    return;
  }
  showCommandFailure(unitId, to, formatLiveCommandError(result), {
    kind: TACTICAL_COMMAND_KIND_SKILL,
    skillId,
  });
}

function runCastSkillFromHud() {
  const selectedId = renderer.selectedUnitId;
  if (selectedId == null || !snapshot) {
    return;
  }
  const caster = snapshot.units.find((u) => u.id === selectedId && u.hp > 0);
  if (!caster) {
    return;
  }
  const target = findStubStrikeTarget(snapshot, caster);
  if (!target) {
    setStatus('射程內無敵方單位，無法施放 stub Strike。', true);
    return;
  }
  void runLiveSkillCommand(selectedId, { x: target.x, y: target.y });
}

async function pollLive() {
  const started = performance.now();
  if (!liveBattleId || !String(liveBattleId).includes('/')) {
    throw new Error('live battle_id missing（請用 /qjp/?live=1 並完成開戰）');
  }
  try {
    const res = await fetch(liveSnapshotUrlForBattle(liveBattleId), {
      cache: 'no-store',
      mode: 'cors',
    });
    if (!res.ok) {
      throw new Error(`live HTTP ${res.status}`);
    }
    const body = await res.json();
    ingestSnapshot(body);
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
    if (pendingCommandFailure) {
      restoreCommandFailureHud();
    } else {
      setStatus('Live Janus mirror polling; legal moves POST to v1/tactical/command.');
      hideLiveRetry();
    }
  } catch (err) {
    syncCtx = {
      source: 'live',
      link: 'error',
      pollIntervalMs: LIVE_POLL_INTERVAL_MS,
    };
    refreshHud();
    const msg = err instanceof Error ? err.message : String(err);
    setStatus(
      `Live 快照失敗（${msg}）。可點「重連 Live」或等待自動重試；同域 v1/… 見 README。`,
      true,
    );
    showLiveRetry('poll', LIVE_LABEL_POLL_RECONNECT, () => {
      void runLivePrepare();
    });
  }
}

async function runMintAccessToken() {
  hideLiveRetry();
  setStatus(`Live：POST ${boot.liveGateway.mintUrlRaw}（暫定 Lares login 鏡像）…`);
  try {
    const token = await mintLiveAccessToken(boot.liveGateway);
    boot.liveGateway.accessToken = token;
    hideMintTokenButton();
    setStatus('Live：已取得 accessToken；開始建局…');
    await runLivePrepare();
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    syncCtx = { source: 'live', link: 'error', pollIntervalMs: LIVE_POLL_INTERVAL_MS };
    refreshHud();
    setStatus(`取 token 失敗：${msg}`, true);
    showMintTokenButton();
    showLiveRetry('prepare', LIVE_LABEL_PREPARE_RETRY, () => {
      void startLiveWithToken();
    });
  }
}

async function ensureLiveAccessToken() {
  if (hasLiveAccessToken(boot.liveGateway)) {
    hideMintTokenButton();
    return true;
  }
  try {
    const token = await mintLiveAccessToken(boot.liveGateway);
    boot.liveGateway.accessToken = token;
    hideMintTokenButton();
    return true;
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    setStatus(`自動 mint 失敗：${msg}。可點「${LIVE_LABEL_MINT_TOKEN}」重試。`, true);
    showMintTokenButton();
    return false;
  }
}

async function startLiveWithToken() {
  syncCtx.source = 'live';
  if (!hasLiveAccessToken(boot.liveGateway)) {
    setStatus('Live：無 accessToken，嘗試同域自動 mint…');
    const ok = await ensureLiveAccessToken();
    if (!ok) {
      return;
    }
  }
  await runLivePrepare();
}

function resetLiveSessionForNewZone(zoneId) {
  boot.liveGateway.zoneId = zoneId;
  persistZoneIdInLocation(zoneId);
  const shard = Number(boot.liveGateway.shard) || 0;
  liveBattleId = `${zoneId}/${shard}`;
  liveSessionId = '';
  stickyTerminalOutcome = null;
  snapshot = null;
  stopLivePoll();
  renderer.clearSelection();
  els.boardWrap?.classList.add('pre-battle');
  campaign?.toDeploy();
  render();
}

function bounceLiveToDeployAfterTerminalZone({ previousZoneId, zoneId, endReason }) {
  resetLiveSessionForNewZone(zoneId);
  const reasonLabel = endReason && endReason !== 'none' ? endReason : '終局';
  setStatus(
    `戰區 ${previousZoneId} 的戰局已結束（${reasonLabel}），已開新戰區 ${zoneId}；請重新排兵佈陣 → 補給 → 開戰。`,
  );
}

async function runLivePrepare() {
  stopLivePoll();
  lastPollAt = 0;
  clearCommandFailure();
  hideLiveRetry();
  if (!hasLiveAccessToken(boot.liveGateway)) {
    setStatus(LIVE_ACCESS_TOKEN_MISSING_MESSAGE, true);
    showMintTokenButton();
    showLiveRetry('prepare', LIVE_LABEL_MINT_TOKEN, () => {
      void runMintAccessToken();
    });
    return;
  }
  setStatus('Live：檢查戰區戰局是否已結束…');
  try {
    const zoneCheck = await ensureFreshLiveZoneBeforeEnter(boot);
    liveBattleId = zoneCheck.battleId;
    if (zoneCheck.action === 'rotated') {
      bounceLiveToDeployAfterTerminalZone(zoneCheck);
      return;
    }
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    setStatus(`Live：戰區檢查失敗（${msg}）；仍嘗試建局…`, true);
  }
  setStatus('Live：POST connect → enter-battle…');
  try {
    const prepared = await prepareLiveJanusSession(boot.liveGateway, liveBattleId);
    liveSessionId = prepared.sessionId;
    liveBattleId = prepared.battleId;
    if (prepared.initialSnapshot && isTerminalLiveSnapshot(prepared.initialSnapshot)) {
      const previousZoneId = boot.liveGateway.zoneId || 'default';
      const zoneId = mintFreshLiveZoneId();
      bounceLiveToDeployAfterTerminalZone({
        previousZoneId,
        zoneId,
        endReason: prepared.initialSnapshot.endReason,
      });
      return;
    }
    if (prepared.initialSnapshot) {
      ingestSnapshot(prepared.initialSnapshot);
      render();
    }
    setStatus(`Live：EnterBattle 已建局 ${liveBattleId}；啟用 auto_command_mode=2…`);
    try {
      await enableLiveAutoCommandModeBoth();
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      setStatus(`Live：auto_command_mode 設定失敗（${msg}），仍嘗試自動迴圈。`, true);
    }
    await pollLive();
    pollTimer = window.setInterval(() => void pollLive(), LIVE_POLL_INTERVAL_MS);
    if (boot.liveAuto) {
      liveAutoPlayer?.start();
      setStatus(
        'Live：自動推進（suggest.command → tactical/command → step-lockstep steps:1；?auto=0 關）。',
      );
    }
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    syncCtx = { source: 'live', link: 'error', pollIntervalMs: LIVE_POLL_INTERVAL_MS };
    refreshHud();
    setStatus(msg, true);
    if (msg.includes(LIVE_ACCESS_TOKEN_MISSING_MESSAGE) || msg.includes('unauthorized')) {
      showMintTokenButton();
    }
    showLiveRetry('prepare', LIVE_LABEL_PREPARE_RETRY, () => {
      void startLiveWithToken();
    });
  }
}

function startLive() {
  void startLiveWithToken();
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
  if (!campaign?.isBattle() && !boot.skipCampaign) {
    return;
  }
  if (!boot.debugMoves) {
    return;
  }
  if (!snapshot || isBattleFinished(snapshot)) {
    return;
  }
  const unitAtCell = unitAt(x, y);

  const selected = renderer.selectedUnitId;
  if (selected != null) {
    if (
      boot.live &&
      !localDrift &&
      unitAtCell &&
      unitAtCell.owner !== LOCAL_PLAYER_OWNER
    ) {
      const caster = snapshot.units.find((u) => u.id === selected && u.hp > 0);
      if (caster && isInAttackRange(caster, x, y)) {
        void runLiveSkillCommand(selected, { x, y });
        return;
      }
    }
    const legal = renderer.legalCells;
    const isLegal = legal.some((c) => c.x === x && c.y === y);
    if (isLegal) {
      if (boot.live && !localDrift) {
        void runLiveMoveCommand(selected, { x, y });
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

  if (boot.live && boot.liveAuto) {
    liveAutoPlayer?.notifyUserAction();
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

function bindIconStrip(container, icons) {
  if (!container) {
    return;
  }
  container.innerHTML = '';
  for (const icon of icons) {
    const span = document.createElement('span');
    span.className = 'res-icon';
    span.title = icon.label;
    const img = document.createElement('img');
    img.src = resolveAppUrl(icon.src);
    img.alt = icon.label;
    img.width = 32;
    img.height = 32;
    span.append(img);
    container.append(span);
  }
}

function bindResourceIcons() {
  bindIconStrip(els.resIcons, HUD_RESOURCE_ICONS);
  bindIconStrip(els.townIcons, HUD_TOWN_ICONS);
}

function decorateCampaignButton(btn, iconSrc) {
  if (!btn || !iconSrc) {
    return;
  }
  btn.classList.add('campaign-btn-icon');
  const img = document.createElement('img');
  img.className = 'campaign-btn-img';
  img.src = resolveAppUrl(iconSrc);
  img.alt = '';
  img.width = 32;
  img.height = 32;
  btn.prepend(img);
}

function bindCampaignButtonIcons() {
  decorateCampaignButton(els.btnCampaignSupply, CAMPAIGN_BUTTON_ICONS.supply);
  decorateCampaignButton(els.btnCampaignBattle, CAMPAIGN_BUTTON_ICONS.battle);
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
    if (card.avatarSrc) {
      const head = document.createElement('div');
      head.className = 'char-card-head';
      const avatar = document.createElement('img');
      avatar.className = 'char-avatar';
      avatar.src = resolveAppUrl(card.avatarSrc);
      avatar.alt = '';
      avatar.width = 64;
      avatar.height = 64;
      avatar.loading = 'lazy';
      const name = document.createElement('span');
      name.textContent = card.label;
      head.append(avatar, name);
      cap.append(head);
    } else {
      cap.textContent = card.label;
    }
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

function beginBattleSession() {
  els.boardWrap?.classList.remove('pre-battle');
  if (boot.live) {
    startLive();
    return;
  }
  loadMock().catch((err) => {
    syncCtx = { source: 'mock', link: 'error' };
    const msg = err instanceof Error ? err.message : String(err);
    setStatus(
      `Mock load failed (${msg}). Serve this folder: cd client/static-preview && npx serve . — or use HTTP, not file://`,
      true,
    );
  });
}

function initCampaign() {
  campaign = createCampaignFlow({
    deployPanel: els.campaignDeploy,
    supplyPanel: els.campaignSupply,
    battlePanel: els.campaignBattle,
    phaseLabel: els.campaignPhase,
    onEnterBattle: beginBattleSession,
  });
  els.btnCampaignSupply?.addEventListener('click', () => campaign.toSupply());
  els.btnCampaignBattle?.addEventListener('click', () => campaign.toBattle());
  els.boardWrap?.classList.add('pre-battle');
  if (boot.skipCampaign) {
    campaign.toBattle();
  }
}

bindResourceIcons();
bindCampaignButtonIcons();
bindCards();

els.mintToken.addEventListener('click', () => {
  void runMintAccessToken();
});

els.castSkill.addEventListener('click', () => {
  runCastSkillFromHud();
});

initCampaign();
if (!boot.skipCampaign) {
  setStatus('請完成排兵佈陣 → 補鎮／補給 → 開戰。');
}

window.addEventListener('beforeunload', () => {
  if (pollTimer) {
    clearInterval(pollTimer);
  }
});
