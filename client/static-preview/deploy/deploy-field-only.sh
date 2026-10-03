#!/usr/bin/env bash
# Deploy ONLY client/static-preview/field/ → ${TARGET_ROOT}/field/
# Does NOT run deploy-web-preview.sh and does NOT rsync the static-preview root.
#
# Safety: verify ${TARGET_ROOT}/index.html SHA-256 unchanged (proves /qjp/ entry not replaced).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
FIELD_SRC="${ROOT}/client/static-preview/field"
TARGET_ROOT="${1:-/var/www/qjp-static-preview}"
TARGET_FIELD="${TARGET_ROOT}/field"
INDEX="${TARGET_ROOT}/index.html"

if [[ ! -d "${FIELD_SRC}" ]]; then
  echo "Missing source: ${FIELD_SRC}" >&2
  exit 1
fi

checksum_index() {
  if [[ -f "${INDEX}" ]]; then
    sha256sum "${INDEX}" | awk '{print $1}'
  else
    echo "(no index.html at target — skip compare)"
  fi
}

BEFORE="$(checksum_index)"
echo "=== field-only deploy ==="
echo "Source:      ${FIELD_SRC}/"
echo "Destination: ${TARGET_FIELD}/"
echo "index.html before: ${BEFORE}"

mkdir -p "${TARGET_FIELD}"
rsync -a \
  --exclude 'README.md' \
  "${FIELD_SRC}/" "${TARGET_FIELD}/"

AFTER="$(checksum_index)"
echo "index.html after:  ${AFTER}"

if [[ "${BEFORE}" != "(no index.html at target — skip compare)" && "${BEFORE}" != "${AFTER}" ]]; then
  echo "ERROR: ${INDEX} checksum changed — abort (root /qjp/ may have been overwritten)." >&2
  exit 2
fi

echo "OK: deployed field/ only; index.html unchanged."
echo "Public URL example: http://47.79.23.223/qjp/field/"
