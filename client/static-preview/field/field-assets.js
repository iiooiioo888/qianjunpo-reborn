/** 手繪稿路徑（僅 field/ 目錄，非 /qjp/ 卡通貼圖） */

export const ASSET_FILES = {
  ground: 'QJP_ground_grass_v01.png',
  infantry: 'QJP_stack_infantry_v01.png',
  archer: 'QJP_stack_archer_v01.png',
  cavalry: 'QJP_stack_cavalry_v01.png',
  panel: 'QJP_panel_teal_gold_v01.png',
};

const cache = /** @type {Record<string, HTMLImageElement>} */ ({});

let loadPromise = null;

export function loadFieldAssets() {
  if (loadPromise) return loadPromise;
  loadPromise = Promise.all(
    Object.entries(ASSET_FILES).map(
      ([key, file]) =>
        new Promise((resolve, reject) => {
          const img = new Image();
          img.onload = () => {
            cache[key] = img;
            resolve();
          };
          img.onerror = () => reject(new Error(`Failed to load ${file}`));
          img.src = file;
        }),
    ),
  );
  return loadPromise;
}

export function getStackImage(type) {
  if (type === 'infantry' || type === 'archer' || type === 'cavalry') {
    return cache[type] ?? null;
  }
  return null;
}

export function getGroundImage() {
  return cache.ground ?? null;
}

export function getPanelImage() {
  return cache.panel ?? null;
}
