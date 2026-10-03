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

/** 與 field-render drawDiamondPath 相同四頂點（top → right → bottom → left） */
export function diamondVertices(topX, topY) {
  return [
    { x: topX, y: topY },
    { x: topX + TILE_W / 2, y: topY + TILE_H / 2 },
    { x: topX, y: topY + TILE_H },
    { x: topX - TILE_W / 2, y: topY + TILE_H / 2 },
  ];
}

function pointInPolygon(px, py, verts) {
  let inside = false;
  for (let i = 0, j = verts.length - 1; i < verts.length; j = i++) {
    const xi = verts[i].x;
    const yi = verts[i].y;
    const xj = verts[j].x;
    const yj = verts[j].y;
    const intersect =
      yi > py !== yj > py && px < ((xj - xi) * (py - yi)) / (yj - yi) + xi;
    if (intersect) inside = !inside;
  }
  return inside;
}

export function pointInDiamond(px, py, topX, topY) {
  return pointInPolygon(px, py, diamondVertices(topX, topY));
}

export function pickGridCell(px, py, originX, originY) {
  for (let gy = 0; gy < GRID_H; gy++) {
    for (let gx = 0; gx < GRID_W; gx++) {
      const { x, y } = gridToScreen(gx, gy, originX, originY);
      if (pointInDiamond(px, py, x, y)) {
        return { x: gx, y: gy };
      }
    }
  }
  return null;
}

/**
 * 點在哪顆畫出來的菱形裡就回該格（不被隔壁兵堆矩形搶走）；
 * 僅在菱形外時，用裁切後兵身矩形補點。
 * @param {{ units: Array<{ id: number, x: number, y: number, troops: number }> }} state
 */
export function pickBattleCell(px, py, originX, originY, state) {
  const cell = pickGridCell(px, py, originX, originY);
  if (cell) {
    return cell;
  }
  const live = state.units.filter((u) => u.troops > 0);
  const drawOrder = [...live].sort((a, b) => b.y + b.x - (a.y + a.x));
  for (const u of drawOrder) {
    const layout = stackLayout(u.x, u.y, originX, originY);
    if (pointInStackPick(px, py, layout)) {
      return { x: u.x, y: u.y };
    }
  }
  return null;
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
