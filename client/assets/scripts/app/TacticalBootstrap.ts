import { _decorator, Component, JsonAsset, Node, resources, Widget } from 'cc';
import { TacticalBoardView } from '../display/TacticalBoardView';
import { TacticalBoardInteraction } from '../display/TacticalBoardInteraction';
import {
  HudSyncLinkState,
  liveLinkStateFromPollAge,
  liveSyncContextFromPoll,
  mockSyncContextFromBootstrap,
} from '../display/LockstepHudFormat';
import { TimeFlowHudStub } from '../display/TimeFlowHudStub';
import { CharacterCardHudStrip } from '../display/CharacterCardHudStrip';
import { resolveCharCardKeyForUnit } from '../display/UnitCharacterCardMapping';
import { ResourceIconHudStrip } from '../display/ResourceIconHudStrip';
import {
  formatTexturePreloadHudNote,
  logTexturePreloadReport,
  preloadTacticalDisplayTextures,
} from '../display/TextureRegistryDev';
import { applyMockMove } from '../logic/MockSnapshotMutator';
import {
  validateSnapshotCellUnitSync,
  validateSnapshotUnitTypesForRegistry,
} from '../logic/MockSnapshotIntegrity';
import { parseViewSnapshot } from '../logic/TacticalSnapshot';
import { DEFAULT_NETWORK_STUB } from '../network/JanusGatewayStub';
import {
  formatJanusLivePrepareError,
  prepareJanusLiveSession,
} from '../network/JanusLiveGatewayHttp';
import { LiveViewSnapshotPoller } from '../network/LiveViewSnapshotPoller';
import { submitTacticalMove } from '../network/TacticalCommandClient';

const { ccclass, property } = _decorator;

@ccclass('TacticalBootstrap')
export class TacticalBootstrap extends Component {
  /** When true, load view JSON from Janus HTTP dev mirror instead of resources mock. */
  @property
  useLiveJanus = false;

  /** Roma battle id (e.g. default/0) from EnterBattle — required when useLiveJanus is true. */
  @property
  liveBattleId = 'default/0';

  /**
   * Live 模式 HTTP 輪詢間隔（ms）。預設 333ms（約 3Hz）。
   * 建議 200–500ms，對齊 Janus `GET /v1/tactical/snapshot` 開發鏡像。
   */
  @property
  livePollIntervalMs = 333;

  /** resources/ relative path without extension (mock mode). */
  @property
  snapshotResource = 'data/tactical/demo_initial';

  /** 僅可選取／移動此 player_id 的單位（與 SubmitTacticalCommand 一致）。 */
  @property
  localPlayerId = 0;

  /**
   * 可選：Janus tactical HTTP 根（如 `http://127.0.0.1:8090`）。
   * 留空 → 瀏覽器預覽用同域相對 `v1/tactical/…`（對齊 Dev `/qjp/v1/` 或 `:18093/v1/`）。
   */
  @property
  janusHttpTacticalBase = '';

  /** Live EnterBattle：`access_token`（與 compose Janus 一致，開發常用 `dev`）。 */
  @property
  liveAccessToken = 'dev';

  @property
  liveZoneId = 'default';

  @property
  liveZoneShard = 0;

  /** 可選：覆寫 HTTP EnterBattle 路徑（預設 `v1/tactical/enter-battle`）。 */
  @property
  janusHttpEnterBattlePath = '';

  /** 可選：覆寫 HTTP Connect 路徑（預設 `v1/tactical/connect`）。 */
  @property
  janusHttpConnectPath = '';

  private poller: LiveViewSnapshotPoller | null = null;
  private liveSessionId = '';
  private boardView: TacticalBoardView | null = null;
  private boardInteraction: TacticalBoardInteraction | null = null;
  private hud: TimeFlowHudStub | null = null;
  private mockSnapshotRaw: unknown | null = null;
  private textureHudNote: string | null = null;
  private lastLiveSnapshotAtMs = 0;
  private livePollAgeTimer: ReturnType<typeof setInterval> | null = null;

  private liveNetworkCfg(): typeof DEFAULT_NETWORK_STUB {
    const base = this.janusHttpTacticalBase.trim();
    const enterPath = this.janusHttpEnterBattlePath.trim();
    const connectPath = this.janusHttpConnectPath.trim();
    return {
      ...DEFAULT_NETWORK_STUB,
      ...(base ? { janusHttpTacticalBase: base } : {}),
      ...(enterPath ? { janusHttpEnterBattlePath: enterPath } : {}),
      ...(connectPath ? { janusHttpConnectPath: connectPath } : {}),
    };
  }

  onDestroy(): void {
    this.poller?.stop();
    this.poller = null;
    if (this.livePollAgeTimer !== null) {
      clearInterval(this.livePollAgeTimer);
      this.livePollAgeTimer = null;
    }
  }

  onLoad(): void {
    const boardNode = new Node('Board');
    boardNode.setParent(this.node);
    const boardView = boardNode.addComponent(TacticalBoardView);
    boardView.cellSize = 32;
    this.boardView = boardView;

    const interaction = boardNode.addComponent(TacticalBoardInteraction);
    interaction.localPlayerId = this.localPlayerId;
    this.boardInteraction = interaction;

    const hudNode = new Node('TimeFlowHud');
    hudNode.setParent(this.node);
    hudNode.setPosition(-320, 300, 0);
    const hud = hudNode.addComponent(TimeFlowHudStub);
    this.hud = hud;
    const hudWidget = hudNode.addComponent(Widget);
    hudWidget.isAlignTop = true;
    hudWidget.isAlignLeft = true;
    hudWidget.top = 16;
    hudWidget.left = 16;

    const iconStripNode = new Node('ResourceIconStrip');
    iconStripNode.setParent(this.node);
    iconStripNode.setPosition(-320, 248, 0);
    const iconStrip = iconStripNode.addComponent(ResourceIconHudStrip);
    const iconWidget = iconStripNode.addComponent(Widget);
    iconWidget.isAlignTop = true;
    iconWidget.isAlignLeft = true;
    iconWidget.top = 72;
    iconWidget.left = 16;

    const charStripNode = new Node('CharacterCardStrip');
    charStripNode.setParent(this.node);
    charStripNode.setPosition(-320, 188, 0);
    const charStrip = charStripNode.addComponent(CharacterCardHudStrip);
    const charWidget = charStripNode.addComponent(Widget);
    charWidget.isAlignTop = true;
    charWidget.isAlignLeft = true;
    charWidget.top = 112;
    charWidget.left = 16;

    const applySnapshot = (raw: unknown, networkStatus: string | null = null) => {
      try {
        const snap = parseViewSnapshot(raw);
        const cellErr = validateSnapshotCellUnitSync(snap);
        if (cellErr) {
          console.warn('[TacticalBootstrap] snapshot cell/unit mismatch', cellErr);
        }
        const typeErr = validateSnapshotUnitTypesForRegistry(snap);
        if (typeErr) {
          console.warn('[TacticalBootstrap] snapshot unit type vs registry', typeErr);
        }
        boardView.applySnapshot(raw);
        const applied = boardView.getSnapshot();
        if (applied) {
          const syncLink: HudSyncLinkState =
            networkStatus != null &&
            (networkStatus.includes('失敗') ||
              networkStatus.includes('Janus:') ||
              networkStatus.includes('拒絕'))
              ? 'error'
              : 'connected';
          if (this.useLiveJanus) {
            this.lastLiveSnapshotAtMs = Date.now();
            hud.updateFromSnapshot(
              applied,
              liveSyncContextFromPoll(0, this.livePollIntervalMs, syncLink),
            );
          } else {
            hud.updateFromSnapshot(applied, mockSyncContextFromBootstrap(syncLink));
          }
          const status =
            networkStatus ??
            (this.textureHudNote && !this.useLiveJanus ? this.textureHudNote : null);
          hud.setNetworkStatus(status);
        }
        this.boardInteraction?.refreshSelectionFromSnapshot();
      } catch (err) {
        console.error('[TacticalBootstrap] snapshot parse/apply failed', err);
        hud.setNetworkStatus('快照解析失敗（見 console）');
      }
    };

    const syncCharCardHighlight = (selectedUnitId: number | null) => {
      const snap = boardView.getSnapshot();
      const unit =
        selectedUnitId != null ? snap?.units.find((u) => u.id === selectedUnitId && u.hp > 0) : undefined;
      if (!unit || unit.owner !== this.localPlayerId) {
        charStrip.setSelectionLinkedCard(null);
        return;
      }
      charStrip.setSelectionLinkedCard(resolveCharCardKeyForUnit(unit));
    };

    interaction.bind(
      boardView,
      async (unitId, to) => {
        const snap = boardView.getSnapshot();
        const unit = snap?.units.find((u) => u.id === unitId);
        if (!unit) {
          return;
        }
        console.info('[TacticalBootstrap] submit move', {
          unitId,
          to,
          playerId: unit.owner,
          battleId: this.liveBattleId,
        });

        if (!this.useLiveJanus) {
          if (this.mockSnapshotRaw == null) {
            hud.setNetworkStatus('Mock：無快照基底');
            return;
          }
          const result = applyMockMove(
            parseViewSnapshot(this.mockSnapshotRaw),
            unitId,
            to,
          );
          if (!result.ok || !result.snapshot) {
            hud.setNetworkStatus(`Mock 移動拒絕：${result.reason ?? 'unknown'}`);
            return;
          }
          this.mockSnapshotRaw = result.snapshot;
          applySnapshot(result.snapshot, 'Mock：已本地套用移動（非權威）');
          this.boardInteraction?.clearSelection();
          return;
        }

        const submit = await submitTacticalMove(this.liveNetworkCfg(), {
          battleId: this.liveBattleId,
          sessionId: this.liveSessionId,
          playerId: unit.owner,
          unitId,
          toX: to.x,
          toY: to.y,
        });
        if (submit.accepted) {
          const hashPart =
            submit.stateHash != null ? ` hash=${submit.stateHash}` : '';
          hud.setNetworkStatus(
            `Live：指令已接受 frame=${submit.lockstepFrame ?? '?'}${hashPart}`,
          );
          this.boardInteraction?.clearSelection();
        } else if (submit.stubOnly) {
          hud.setNetworkStatus(submit.rejectReason ?? 'Live：HTTP 指令鏡像未部署');
          console.info('[TacticalBootstrap] live submit stub — use grpcurl SubmitTacticalCommand');
        } else {
          hud.setNetworkStatus(`Live 拒絕：${submit.rejectReason ?? 'unknown'}`);
        }
      },
      syncCharCardHighlight,
    );

    const displayMode = this.useLiveJanus ? 'live' : 'mock';
    void preloadTacticalDisplayTextures().then((report) => {
      logTexturePreloadReport(report, displayMode);
      this.textureHudNote = formatTexturePreloadHudNote(report);
      iconStrip.buildStrip();
      charStrip.buildStrip();
      if (this.useLiveJanus) {
        const startLivePoller = () => {
          this.poller = new LiveViewSnapshotPoller({
            cfg: this.liveNetworkCfg(),
            battleId: this.liveBattleId,
            sessionId: this.liveSessionId,
            intervalMs: this.livePollIntervalMs,
            onSnapshot: (raw) => applySnapshot(raw, null),
            onError: (err, retryMs) => {
              console.warn(
                '[TacticalBootstrap] live Janus poll failed',
                err.message,
                `(retry ~${retryMs}ms)`,
              );
              hud.setNetworkStatus(`Janus: ${err.message}（約 ${retryMs}ms 後重試）`);
              const age =
                this.lastLiveSnapshotAtMs > 0 ? Date.now() - this.lastLiveSnapshotAtMs : retryMs;
              hud.setLockstepSyncContext(
                liveSyncContextFromPoll(age, this.livePollIntervalMs, 'error'),
              );
            },
          });
          this.poller.start(true);
          if (this.livePollAgeTimer !== null) {
            clearInterval(this.livePollAgeTimer);
          }
          this.livePollAgeTimer = setInterval(() => {
            if (!this.useLiveJanus || this.lastLiveSnapshotAtMs <= 0) {
              return;
            }
            const snap = this.boardView?.getSnapshot();
            if (!snap || !this.hud) {
              return;
            }
            const age = Date.now() - this.lastLiveSnapshotAtMs;
            const link = liveLinkStateFromPollAge(age, this.livePollIntervalMs);
            this.hud.setLockstepSyncContext(
              liveSyncContextFromPoll(age, this.livePollIntervalMs, link),
            );
            this.hud.updateFromSnapshot(snap);
          }, 250);
        };

        void (async () => {
          hud.setNetworkStatus('Live：POST connect → enter-battle…');
          try {
            const prepared = await prepareJanusLiveSession({
              cfg: this.liveNetworkCfg(),
              accessToken: this.liveAccessToken,
              clientVersion: DEFAULT_NETWORK_STUB.clientVersion,
              targetZone: { zoneId: this.liveZoneId, shard: this.liveZoneShard },
              battleIdHint: this.liveBattleId,
            });
            this.liveSessionId = prepared.sessionId;
            this.liveBattleId = prepared.battleId;
            if (prepared.initialSnapshot) {
              applySnapshot(prepared.initialSnapshot, `Live：EnterBattle 已建局 ${prepared.battleId}`);
            } else {
              hud.setNetworkStatus(`Live：EnterBattle 已建局 ${prepared.battleId}（等待快照）`);
            }
            startLivePoller();
          } catch (err) {
            const msg = formatJanusLivePrepareError(err);
            console.error('[TacticalBootstrap] live session prepare failed', err);
            hud.setNetworkStatus(msg);
            hud.setLockstepSyncContext(
              liveSyncContextFromPoll(0, this.livePollIntervalMs, 'error'),
            );
          }
        })();
        return;
      }

      resources.load(this.snapshotResource, JsonAsset, (err, asset) => {
        if (err || !asset) {
          console.error('[TacticalBootstrap] failed to load mock snapshot', err);
          hud.setNetworkStatus('Mock JSON 載入失敗');
          return;
        }
        this.mockSnapshotRaw = asset.json;
        applySnapshot(asset.json);
      });
    });
  }
}
