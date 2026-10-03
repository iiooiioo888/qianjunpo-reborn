/** 斜角格座標 ↔ 畫布像素（自製菱形，非 ISO25 貼圖） */

import { pointInStackPick, stackLayout } from './field-stack-layout.js';

export const GRID_W = 13;
export const GRID_H = 13;
export const TILE_W = 48;
export const TILE_H = 24;

export function gridToScreen(gx, gy, originX, originY) {
  const x = originX + (gx - gy) * (TILE_W / 2);
  const y = originY + (gx + gy) * (TILE_H / 2);
  return { x, y };
}

export function screenToGrid(px, py, originX, originY) {
  const rx = px - originX;
  const ry = py - originY;
  const halfH = TILE_H / 2;
  const gx = (rx / (TILE_W / 2) + ry / halfH) / 2;
  const gy = (ry / halfH - rx / (TILE_W / 2)) / 2;
  return { gx, gy };
}

export function pickGridCell(px, py, originX, originY) {
  const { gx, gy } = screenToGrid(px, py, originX, originY);
  const cx = Math.round(gx);
  const cy = Math.round(gy);
  if (cx < 0 || cy < 0 || cx >= GRID_W || cy >= GRID_H) {
    return null;
  }
  const center = gridToScreen(cx, cy, originX, originY);
  const dx = Math.abs(gx - cx) + Math.abs(gy - cy);
  if (dx > 0.55) {
    return null;
  }
  const dist = Math.hypot(px - center.x, py - (center.y + TILE_H / 2));
  if (dist > TILE_W * 0.55) {
    return null;
  }
  return { x: cx, y: cy };
}

/**
 * 先命中兵堆圖（含高出菱形的上半部），否則走菱形空格判定。
 * @param {{ units: Array<{ id: number, x: number, y: number, troops: number }> }} state
 */
export function pickBattleCell(px, py, originX, originY, state) {
  const live = state.units.filter((u) => u.troops > 0);
  const drawOrder = [...live].sort((a, b) => b.y + b.x - (a.y + a.x));
  for (const u of drawOrder) {
    const layout = stackLayout(u.x, u.y, originX, originY);
    if (pointInStackPick(px, py, layout)) {
      return { x: u.x, y: u.y };
    }
  }
  return pickGridCell(px, py, originX, originY);
}

export function computeOrigin(canvasWidth, canvasHeight) {
  const mapW = (GRID_W + GRID_H) * (TILE_W / 2);
  const mapH = (GRID_W + GRID_H) * (TILE_H / 2) + TILE_H;
  const originX = (canvasWidth - mapW) / 2 + (GRID_H * TILE_W) / 2;
  const originY = (canvasHeight - mapH) / 2 + TILE_H;
  return { originX, originY };
}

export function manhattan(a, b) {
  return Math.abs(a.x - b.x) + Math.abs(a.y - b.y);
}
