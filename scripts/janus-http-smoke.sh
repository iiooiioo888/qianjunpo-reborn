#!/usr/bin/env bash
# Janus HTTP tactical smoke: connect → enter-battle → command (Compose / dev).
# Requires: curl, jq (or python3). Set ACCESS_TOKEN (Lares-signed; see docs/janus-http-mirror.md).
# Usage: ACCESS_TOKEN='…' ./scripts/janus-http-smoke.sh
#        JANUS_HTTP_BASE=http://127.0.0.1:18090 ACCESS_TOKEN='…' ./scripts/janus-http-smoke.sh
set -euo pipefail

JANUS_HTTP_BASE="${JANUS_HTTP_BASE:-http://127.0.0.1:18090}"
PLAYER_ID="${PLAYER_ID:-0}"
MOVE_KIND="${MOVE_KIND:-1}"
TO_X="${TO_X:-5}"
TO_Y="${TO_Y:-8}"
FALLBACK_UNIT_ID="${FALLBACK_UNIT_ID:-101}"
BASE="${JANUS_HTTP_BASE%/}"

if ! command -v curl >/dev/null 2>&1; then
  echo "error: curl is required" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1 && ! command -v python3 >/dev/null 2>&1; then
  echo "error: jq or python3 is required" >&2
  exit 1
fi

json_field() {
  local json="$1"
  local key="$2"
  if command -v jq >/dev/null 2>&1; then
    echo "$json" | jq -r --arg k "$key" '.[$k] // empty'
  else
    echo "$json" | python3 -c 'import json,sys; k=sys.argv[1]; d=json.load(sys.stdin); v=d.get(k); print("" if v is None else v)' "$key"
  fi
}

pick_unit_id() {
  local snapshot_json="$1"
  if command -v jq >/dev/null 2>&1; then
    echo "$snapshot_json" | jq -r --argjson owner "$PLAYER_ID" '
      (.units // []) | map(select(.owner == $owner)) | .[0].id // empty
    '
  else
    echo "$snapshot_json" | python3 -c "
import json, sys, os
owner = int(os.environ.get('PLAYER_ID', '0'))
data = json.load(sys.stdin)
for u in data.get('units') or []:
    if int(u.get('owner', -1)) == owner and 'id' in u:
        print(u['id'])
        break
"
  fi
}

mint_access_token() {
  local mint_body
  mint_body=$(curl -sS -X POST "${BASE}/v1/auth/dev-mint" \
    -H 'Content-Type: application/json' -d '{}' || true)
  json_field "$mint_body" access_token
}

if [[ -z "${ACCESS_TOKEN:-}" ]]; then
  ACCESS_TOKEN=$(mint_access_token || true)
  if [[ -z "$ACCESS_TOKEN" ]]; then
    echo "error: ACCESS_TOKEN is required (set ACCESS_TOKEN, enable JANUS_HTTP_DEV_MINT, or Lares Login; see docs/janus-http-mirror.md)" >&2
    exit 1
  fi
  echo "==> minted ACCESS_TOKEN via POST ${BASE}/v1/auth/dev-mint"
fi

echo "==> POST ${BASE}/v1/tactical/connect"
CONNECT_BODY=$(curl -sS -X POST "${BASE}/v1/tactical/connect" \
  -H 'Content-Type: application/json' \
  -d "{\"access_token\":\"${ACCESS_TOKEN}\",\"target_zone\":{\"zone_id\":\"default\",\"shard\":0}}")
SESSION_ID=$(json_field "$CONNECT_BODY" session_id)
if [[ -z "$SESSION_ID" ]]; then
  echo "connect failed: $CONNECT_BODY" >&2
  exit 1
fi

echo "==> POST ${BASE}/v1/tactical/enter-battle"
ENTER_BODY=$(curl -sS -X POST "${BASE}/v1/tactical/enter-battle" \
  -H 'Content-Type: application/json' \
  -d "{\"session_id\":\"${SESSION_ID}\",\"access_token\":\"${ACCESS_TOKEN}\",\"target_zone\":{\"zone_id\":\"default\",\"shard\":0}}")
BATTLE_ID=$(json_field "$ENTER_BODY" battle_id)
if [[ -z "$BATTLE_ID" ]]; then
  echo "enter-battle failed: $ENTER_BODY" >&2
  exit 1
fi

if command -v jq >/dev/null 2>&1; then
  SNAPSHOT_JSON=$(echo "$ENTER_BODY" | jq -c '.view_snapshot_json // {}')
else
  SNAPSHOT_JSON=$(echo "$ENTER_BODY" | python3 -c 'import json,sys; print(json.dumps(json.load(sys.stdin).get("view_snapshot_json") or {}))')
fi

UNIT_ID=$(pick_unit_id "$SNAPSHOT_JSON")
if [[ -z "$UNIT_ID" ]]; then
  UNIT_ID="$FALLBACK_UNIT_ID"
fi

echo "==> POST ${BASE}/v1/tactical/command (unit_id=${UNIT_ID})"
CMD_BODY=$(curl -sS -X POST "${BASE}/v1/tactical/command" \
  -H 'Content-Type: application/json' \
  -d "{\"session_id\":\"${SESSION_ID}\",\"battle_id\":\"${BATTLE_ID}\",\"player_id\":${PLAYER_ID},\"kind\":${MOVE_KIND},\"unit_id\":${UNIT_ID},\"to_x\":${TO_X},\"to_y\":${TO_Y}}")

ACCEPTED=$(json_field "$CMD_BODY" accepted)
if [[ "$ACCEPTED" != "True" && "$ACCEPTED" != "true" ]]; then
  echo "command not accepted: $CMD_BODY" >&2
  exit 1
fi

echo "ok: battle_id=${BATTLE_ID} unit_id=${UNIT_ID} accepted=true"
