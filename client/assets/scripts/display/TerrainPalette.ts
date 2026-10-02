import { Color } from 'cc';
import { TerrainKind } from '../logic/TacticalSnapshot';

/** Flat fills when ISO25 tile sprites are missing (Nearest + integer scale). */
export function terrainFillColor(t: TerrainKind): Color {
  switch (t) {
    case TerrainKind.Mountain:
      return new Color(120, 110, 100, 255);
    case TerrainKind.Forest:
      return new Color(56, 140, 72, 255);
    case TerrainKind.River:
      return new Color(48, 96, 180, 255);
    case TerrainKind.City:
      return new Color(180, 160, 120, 255);
    case TerrainKind.Pass:
      return new Color(200, 180, 90, 255);
    case TerrainKind.Plain:
    default:
      return new Color(92, 148, 72, 255);
  }
}

export function ownerAccentColor(owner: number): Color {
  return owner === 0 ? new Color(220, 80, 70, 255) : new Color(70, 120, 220, 255);
}
