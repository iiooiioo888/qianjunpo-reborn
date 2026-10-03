#!/usr/bin/env bash
# Pack static-preview + synced textures for /var/www/qjp-static-preview (see nginx-qjp-snippet.conf).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PREVIEW="${ROOT}/client/static-preview"
TARGET="${1:-/var/www/qjp-static-preview}"

bash "${PREVIEW}/scripts/sync-textures-from-assets.sh"

mkdir -p "${TARGET}"
rsync -a --delete \
  --exclude 'deploy/' \
  --exclude 'scripts/' \
  --exclude 'README.md' \
  "${PREVIEW}/" "${TARGET}/"

echo "Deployed static-preview -> ${TARGET}"
echo "Ensure nginx uses client/static-preview/deploy/nginx-qjp-snippet.conf (301 must keep query: \$is_args\$args)."
