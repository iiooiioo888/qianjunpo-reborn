/**
 * Paths under static-preview/textures/2d/ (synced from client/assets/resources/textures/2d).
 * Aligned with Cocos *SpriteRegistry stems.
 */

export const TEXTURE_2D_BASE = 'textures/2d';

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

export const CHAR_CARD_V04 = [
  { label: '曹操', src: `${TEXTURE_2D_BASE}/chars/PX2D_CHAR_WEI_Caocao_ex_v04.png` },
  { label: '張飛', src: `${TEXTURE_2D_BASE}/chars/PX2D_CHAR_SHU_Zhangfei_ex_v04.png` },
  { label: '吳', src: `${TEXTURE_2D_BASE}/chars/PX2D_CHAR_WU_Placeholder_01_v04.png` },
];

/** HUD resource strip (32px art, integer upscale in CSS). */
export const HUD_RESOURCE_ICONS = [
  { label: '糧', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_res_food_32.png` },
  { label: '木', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_res_wood_32.png` },
  { label: '石', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_res_stone_32.png` },
  { label: '鐵', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_res_iron_32.png` },
  { label: '金', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_res_gold_32.png` },
  { label: '軍', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_army_32.png` },
];

/** 城鎮／建築 HUD（32px，補鎮步驟展示） */
export const HUD_TOWN_ICONS = [
  { label: '主城', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_bld_keep_32.png` },
  { label: '兵營', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_bld_barracks_32.png` },
  { label: '倉庫', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_bld_warehouse_32.png` },
  { label: '學堂', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_bld_academy_32.png` },
  { label: '伐木', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_bld_lumber_32.png` },
  { label: '陣形', src: `${TEXTURE_2D_BASE}/icons/PX2D_icon_formation_32.png` },
];

export const UNIT_ART_CANVAS_PX = 128;
