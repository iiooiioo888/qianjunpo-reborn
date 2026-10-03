/** 主線：排兵佈陣 → 補鎮／補給 → 開戰 */

export const PHASE_DEPLOY = 'deploy';
export const PHASE_SUPPLY = 'supply';
export const PHASE_BATTLE = 'battle';

const PHASE_LABEL = {
  [PHASE_DEPLOY]: '① 排兵佈陣',
  [PHASE_SUPPLY]: '② 補鎮／補給',
  [PHASE_BATTLE]: '③ 開戰',
};

/**
 * @param {{ deployPanel: HTMLElement, supplyPanel: HTMLElement, battlePanel: HTMLElement, phaseLabel: HTMLElement, onEnterBattle: () => void }} opts
 */
export function createCampaignFlow(opts) {
  let phase = PHASE_DEPLOY;

  function syncUi() {
    opts.phaseLabel.textContent = PHASE_LABEL[phase] ?? phase;
    opts.deployPanel.hidden = phase !== PHASE_DEPLOY;
    opts.supplyPanel.hidden = phase !== PHASE_SUPPLY;
    opts.battlePanel.hidden = phase !== PHASE_BATTLE;
  }

  function toSupply() {
    phase = PHASE_SUPPLY;
    syncUi();
  }

  function toDeploy() {
    phase = PHASE_DEPLOY;
    syncUi();
  }

  function toBattle() {
    phase = PHASE_BATTLE;
    syncUi();
    opts.onEnterBattle();
  }

  syncUi();

  return {
    getPhase: () => phase,
    isBattle: () => phase === PHASE_BATTLE,
    toDeploy,
    toSupply,
    toBattle,
  };
}
