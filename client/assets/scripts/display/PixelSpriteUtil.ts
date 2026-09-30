import { SpriteFrame, Texture2D } from 'cc';

/** Cartoon pixel风：Nearest / Point，關閉 mipmap 語意（Cocos 2D 預設無 mipmap）。 */
export function applyPixelArtSampling(frame: SpriteFrame | null): void {
  if (!frame?.texture) {
    return;
  }
  const tex = frame.texture;
  tex.setFilters(Texture2D.Filter.NEAREST, Texture2D.Filter.NEAREST);
}
