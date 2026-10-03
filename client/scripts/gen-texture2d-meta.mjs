#!/usr/bin/env node
/**
 * Generate Cocos Creator 3.8 image .meta for resources/textures/2d PNGs
 * (filterMode nearest, sprite-frame default). Idempotent UUIDs from relative path.
 */
import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync, readdirSync, statSync } from 'node:fs';
import { dirname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const texRoot = join(root, 'assets/resources/textures/2d');

function uuidFromSeed(seed) {
  const h = createHash('sha256').update(seed).digest('hex');
  return `${h.slice(0, 8)}-${h.slice(8, 12)}-${h.slice(12, 16)}-${h.slice(16, 20)}-${h.slice(20, 32)}`;
}

function pngSize(buf) {
  if (buf.length < 24 || buf.readUInt32BE(0) !== 0x89504e47) {
    return { width: 64, height: 64 };
  }
  return { width: buf.readUInt32BE(16), height: buf.readUInt32BE(20) };
}

function walkPngs(dir, out = []) {
  for (const name of readdirSync(dir)) {
    const p = join(dir, name);
    if (statSync(p).isDirectory()) {
      walkPngs(p, out);
    } else if (name.endsWith('.png')) {
      out.push(p);
    }
  }
  return out;
}

function buildMeta(relPath, width, height) {
  const imageUuid = uuidFromSeed(`img:${relPath}`);
  const texId = '6c48a';
  const sfId = 'f9941';
  const texUuid = uuidFromSeed(`tex:${relPath}`);
  const sfUuid = uuidFromSeed(`sf:${relPath}`);
  const baseName = relPath.split('/').pop().replace(/\.png$/, '');
  return {
    ver: '1.0.27',
    importer: 'image',
    imported: true,
    uuid: imageUuid,
    files: ['.json', '.png'],
    subMetas: {
      [texId]: {
        importer: 'texture',
        uuid: texUuid,
        displayName: baseName,
        id: texId,
        name: 'texture',
        userData: {
          wrapModeS: 'clamp-to-edge',
          wrapModeT: 'clamp-to-edge',
          minfilter: 'nearest',
          magfilter: 'nearest',
          mipfilter: 'none',
          anisotropy: 0,
          isUuid: true,
          imageUuidOrDatabaseUri: imageUuid,
        },
        ver: '1.0.22',
        imported: true,
        files: ['.json'],
        subMetas: {},
      },
      [sfId]: {
        importer: 'sprite-frame',
        uuid: sfUuid,
        displayName: baseName,
        id: sfId,
        name: 'spriteFrame',
        userData: {
          trimType: 'auto',
          trimThreshold: 1,
          rotated: false,
          offsetX: 0,
          offsetY: 0,
          trimX: 0,
          trimY: 0,
          width,
          height,
          rawWidth: width,
          rawHeight: height,
          borderTop: 0,
          borderBottom: 0,
          borderLeft: 0,
          borderRight: 0,
          packable: true,
          pixelsToUnit: 100,
          pivotX: 0.5,
          pivotY: 0.5,
          meshType: 0,
          isUuid: true,
          imageUuidOrDatabaseUri: imageUuid,
          atlasUuid: '',
        },
        ver: '1.0.12',
        imported: true,
        files: ['.json'],
        subMetas: {},
      },
    },
    userData: {
      type: 'sprite-frame',
      hasAlpha: true,
      fixAlphaTransparencyArtifacts: false,
      redirect: sfId,
      flipGreenChannel: false,
      flipVertical: false,
    },
  };
}

const pngs = walkPngs(texRoot);
let wrote = 0;
for (const abs of pngs) {
  const rel = relative(texRoot, abs).replace(/\\/g, '/');
  const buf = readFileSync(abs);
  const { width, height } = pngSize(buf);
  const metaPath = `${abs}.meta`;
  const meta = buildMeta(rel, width, height);
  writeFileSync(metaPath, `${JSON.stringify(meta, null, 2)}\n`);
  wrote += 1;
}
console.log(`wrote ${wrote} .meta under ${texRoot}`);
