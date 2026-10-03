/**
 * static-preview 資產路徑（與 Cocos stem 對齊；deploy 前跑 sync-textures-from-assets.sh）。
 */

export const TEXTURE_2D_BASE = 'textures/2d';

/** TerrainKind 0–3（城 4／關 5 待 art PR 另開，本 PR 不掛）。 */
export const TERRAIN_TILE_SRC = {
  0: `${TEXTURE_2D_BASE}/tiles/ISO25_tile_grass_v02.png`,
  1: `${TEXTURE_2D_BASE}/tiles/ISO25_tile_mountain_v01.png`,
  2: `${TEXTURE_2D_BASE}/tiles/ISO25_tile_forest_v01.png`,
  3: `${TEXTURE_2D_BASE}/tiles/ISO25_tile_water_v01.png`,
};

export const UNIT_TEXTURE_SRC = {
  0: `${TEXTURE_2D_BASE}/units/PX2D_unit_infantry.png`,
  1: `${TEXTURE_2D_BASE}/units/PX2D_unit_infantry.png`,
  2: `${TEXTURE_2D_BASE}/units/PX2D_unit_cavalry.png`,
};

/** 角色卡：static-preview/chars/（config CHAR_CARD_V04 路徑） */
export const CHAR_CARD_V04 = [
  { label: '曹操', src: 'chars/PX2D_CHAR_WEI_Caocao_ex_v04.png' },
  { label: '張飛', src: 'chars/PX2D_CHAR_SHU_Zhangfei_ex_v04.png' },
  { label: '吳', src: 'chars/PX2D_CHAR_WU_Placeholder_01_v04.png' },
];

export const HUD_RESOURCE_ICONS = [
  { label: '糧', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_res_food_32.png` },
  { label: '金', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_res_gold_32.png` },
  { label: '兵', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_army_32.png` },
];

export const HUD_TOWN_ICONS = [
  { label: '城寨', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_bld_keep_32.png` },
  { label: '糧倉', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_bld_warehouse_32.png` },
];

export const UNIT_ART_CANVAS_PX = 128;
