import { ViewUnit } from '../logic/TacticalSnapshot';
import { CHAR_CARD_TEXTURE_KEYS, CharCardTextureKey } from './CharacterCardSpriteRegistry';

/**
 * Mock 選取→角色卡條聯動（display stub，非權威資料）。
 * 解析順序：① `unitId` 覆寫 → ② `type` 預設 → ③ 吳占位卡。
 *
 * | unitId (demo) | type | 卡鍵 |
 * |---------------|------|------|
 * | 101           | 0    | char_caocao |
 * | (其他)        | 0    | char_caocao |
 * | (其他)        | 2    | char_zhangfei |
 */
const MOCK_UNIT_ID_TO_CARD: Readonly<Partial<Record<number, CharCardTextureKey>>> = {
  101: CHAR_CARD_TEXTURE_KEYS.char_caocao,
};

const MOCK_UNIT_TYPE_TO_CARD: Readonly<Partial<Record<number, CharCardTextureKey>>> = {
  0: CHAR_CARD_TEXTURE_KEYS.char_caocao,
  2: CHAR_CARD_TEXTURE_KEYS.char_zhangfei,
};

export function resolveCharCardKeyForUnit(unit: ViewUnit | null | undefined): CharCardTextureKey | null {
  if (!unit) {
    return null;
  }
  const byId = MOCK_UNIT_ID_TO_CARD[unit.id];
  if (byId) {
    return byId;
  }
  const byType = MOCK_UNIT_TYPE_TO_CARD[unit.type];
  if (byType) {
    return byType;
  }
  return CHAR_CARD_TEXTURE_KEYS.char_wu_placeholder;
}
