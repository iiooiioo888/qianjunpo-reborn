#!/usr/bin/env bash
# Janus HTTP tactical smoke: connect → enter-battle → command (Compose / dev).
# Requires: curl, jq (or python3). Set ACCESS_TOKEN (Lares-signed; see docs/janus-http-mirror.md).
# Usage: ACCESS_TOKEN='…' ./scripts/janus-http-smoke.sh
#        JANUS_HTTP_BASE=http://127.0.0.1:18090 ACCESS_TOKEN='…' ./scripts/janus-http-smoke.sh
# KindSkill E2E (bridge move if needed, then cast):
#   MOVE_KIND=5 SKILL_ID=1 ./scripts/janus-http-smoke.sh
#   SKILL_ID=1 ./scripts/janus-http-smoke.sh   # sets kind=5; targets default-duel enemy @ (16,10)
# After a finished match, assert outcome on snapshot JSON (rebuild Compose images if schema changed):
#   curl -sS "${JANUS_HTTP_BASE}/v1/tactical/snapshot?battle_id=…" | jq '{finished,winner,endReason}'
set -euo pipefail

JANUS_HTTP_BASE="${JANUS_HTTP_BASE:-http://127.0.0.1:18090}"
PLAYER_ID="${PLAYER_ID:-0}"
MOVE_KIND="${MOVE_KIND:-1}"
SKILL_ID="${SKILL_ID:-}"
TO_X="${TO_X:-5}"
TO_Y="${TO_Y:-8}"
FALLBACK_UNIT_ID="${FALLBACK_UNIT_ID:-101}"
# Lockstep frames before a command resolves (pkg/lockstep.CommandDelayFrames + 1).
LOCKSTEP_STEPS="${LOCKSTEP_STEPS:-4}"
MAX_BRIDGE_MOVES="${MAX_BRIDGE_MOVES:-24}"
BASE="${JANUS_HTTP_BASE%/}"

SKILL_MODE=0
if [[ -n "$SKILL_ID" ]] || [[ "$MOVE_KIND" == "5" ]]; then
  SKILL_MODE=1
  MOVE_KIND=5
  if [[ -z "$SKILL_ID" ]]; then
    SKILL_ID=1
  fi
fi

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
  mint_body=$(curl -sS -X POST "${BASE}/v1/lares/login" \
    -H 'Content-Type: application/json' \
    -d '{"username":"smoke","password":"smoke"}' || true)
  json_field "$mint_body" access_token
}

post_command() {
  local kind="$1"
  local unit_id="$2"
  local to_x="$3"
  local to_y="$4"
  local skill_id="${5:-}"
  local cmd_json
  cmd_json="{\"session_id\":\"${SESSION_ID}\",\"battle_id\":\"${BATTLE_ID}\",\"player_id\":${PLAYER_ID},\"kind\":${kind},\"unit_id\":${unit_id},\"to_x\":${to_x},\"to_y\":${to_y}"
  if [[ -n "$skill_id" ]]; then
    cmd_json="${cmd_json},\"skill_id\":${skill_id}"
  fi
  cmd_json="${cmd_json}}"
  curl -sS -X POST "${BASE}/v1/tactical/command" \
    -H 'Content-Type: application/json' \
    -d "$cmd_json"
}

assert_command_accepted() {
  local label="$1"
  local body="$2"
  local accepted
  accepted=$(json_field "$body" accepted)
  if [[ "$accepted" != "True" && "$accepted" != "true" ]]; then
    echo "${label} not accepted: $body" >&2
    exit 1
  fi
}

step_lockstep() {
  local steps="${1:-$LOCKSTEP_STEPS}"
  curl -sS -X POST "${BASE}/v1/tactical/step-lockstep" \
    -H 'Content-Type: application/json' \
    -d "{\"session_id\":\"${SESSION_ID}\",\"battle_id\":\"${BATTLE_ID}\",\"steps\":${steps}}"
}

fetch_snapshot() {
  curl -sS "${BASE}/v1/tactical/snapshot?battle_id=${BATTLE_ID}"
}

# Prints: skill_to_x skill_to_y bridge_x bridge_y
plan_skill_bridge() {
  local snapshot_json="$1"
  local unit_id="$2"
  echo "$snapshot_json" | PLAYER_ID="$PLAYER_ID" \
    SKILL_TO_X="${SKILL_TO_X:-}" SKILL_TO_Y="${SKILL_TO_Y:-}" \
    BRIDGE_TO_X="${BRIDGE_TO_X:-}" BRIDGE_TO_Y="${BRIDGE_TO_Y:-}" \
    python3 -c '
import json, os, sys
owner = int(os.environ.get("PLAYER_ID", "0"))
uid = int(sys.argv[1])
data = json.load(sys.stdin)
units = data.get("units") or []
board_size = int(data.get("boardSize") or 20)
cells = data.get("cells") or []

def cheb(a, b):
    return max(abs(a[0]-b[0]), abs(a[1]-b[1]))

def passable(x, y):
    if y < 0 or y >= len(cells) or x < 0 or x >= len(cells[y]):
        return True
    return bool(cells[y][x].get("passable", True))

def occupied(x, y):
    return any(u.get("x") == x and u.get("y") == y and int(u.get("id")) != uid for u in units)

self_u = next((u for u in units if int(u.get("id")) == uid), None)
if not self_u:
    raise SystemExit("caster unit not in snapshot")
otx, oty = os.environ.get("SKILL_TO_X", ""), os.environ.get("SKILL_TO_Y", "")
if otx and oty:
    tx, ty = int(otx), int(oty)
else:
    enemy = next((u for u in units if int(u.get("owner", -1)) != owner), None)
    if not enemy:
        raise SystemExit("no enemy unit in snapshot")
    tx, ty = int(enemy["x"]), int(enemy["y"])
obx, oby = os.environ.get("BRIDGE_TO_X", ""), os.environ.get("BRIDGE_TO_Y", "")
if obx and oby:
    bx, by = int(obx), int(oby)
else:
    candidates = []
    for dx, dy in [(-1, 0), (1, 0), (0, -1), (0, 1), (-1, -1), (-1, 1), (1, -1), (1, 1)]:
        x, y = tx + dx, ty + dy
        if 0 <= x < board_size and 0 <= y < board_size and passable(x, y) and not occupied(x, y):
            candidates.append([x, y])
    if not candidates:
        raise SystemExit("no bridge cell adjacent to skill target")
    sx, sy = int(self_u["x"]), int(self_u["y"])
    bridge = min(candidates, key=lambda c: cheb([sx, sy], c))
    bx, by = bridge[0], bridge[1]
print(f"{tx} {ty} {bx} {by}")
' "$unit_id"
}

# Prints next move destination toward bridge (within MOVE_POINTS along BFS path).
next_bridge_move_dest() {
  local snapshot_json="$1"
  local unit_id="$2"
  local bridge_x="$3"
  local bridge_y="$4"
  echo "$snapshot_json" | MOVE_POINTS="${MOVE_POINTS:-4}" python3 -c '
import json, os, sys
from collections import deque

uid = int(sys.argv[1])
bx, by = int(sys.argv[2]), int(sys.argv[3])
move_points = int(os.environ.get("MOVE_POINTS", "4"))
data = json.load(sys.stdin)
units = data.get("units") or []
board_size = int(data.get("boardSize") or 20)
cells = data.get("cells") or []

def passable(x, y):
    if y < 0 or y >= len(cells) or x < 0 or x >= len(cells[y]):
        return True
    return bool(cells[y][x].get("passable", True))

def unit_at(x, y):
    for u in units:
        if int(u.get("x")) == x and int(u.get("y")) == y:
            return int(u.get("id"))
    return 0

self_u = next((u for u in units if int(u.get("id")) == uid), None)
if not self_u:
    raise SystemExit("caster unit not in snapshot")
start = (int(self_u["x"]), int(self_u["y"]))
goal = (bx, by)
if start == goal:
    print(f"{bx} {by}")
    raise SystemExit(0)

def neighbors(x, y):
    for dx in (-1, 0, 1):
        for dy in (-1, 0, 1):
            if dx == 0 and dy == 0:
                continue
            nx, ny = x + dx, y + dy
            if 0 <= nx < board_size and 0 <= ny < board_size:
                yield nx, ny

prev = {start: None}
q = deque([start])
found = None
while q:
    cur = q.popleft()
    if cur == goal:
        found = cur
        break
    for nx, ny in neighbors(*cur):
        if (nx, ny) in prev:
            continue
        if not passable(nx, ny):
            continue
        occ = unit_at(nx, ny)
        if occ not in (0, uid):
            continue
        prev[(nx, ny)] = cur
        q.append((nx, ny))

if found is None:
    raise SystemExit(f"no path to bridge ({bx},{by}) from {start}")

path = []
cur = goal
while cur is not None:
    path.append(cur)
    cur = prev[cur]
path.reverse()
step_idx = min(move_points, len(path) - 1)
dest = path[step_idx]
print(f"{dest[0]} {dest[1]}")
' "$unit_id" "$bridge_x" "$bridge_y"
}

unit_cheb_to_target() {
  local snapshot_json="$1"
  local unit_id="$2"
  local tx="$3"
  local ty="$4"
  if command -v jq >/dev/null 2>&1; then
    echo "$snapshot_json" | jq -r \
      --argjson uid "$unit_id" \
      --argjson tx "$tx" \
      --argjson ty "$ty" '
      (.units // []) | map(select(.id == $uid)) | .[0] as $u
      | if $u == null then 999
        else [($u.x - $tx | fabs), ($u.y - $ty | fabs)] | max
        end
    '
  else
    echo "$snapshot_json" | python3 -c '
import json, sys
uid, tx, ty = int(sys.argv[1]), int(sys.argv[2]), int(sys.argv[3])
data = json.load(sys.stdin)
for u in data.get("units") or []:
    if int(u.get("id")) == uid:
        print(max(abs(int(u["x"]) - tx), abs(int(u["y"]) - ty)))
        break
else:
    print(999)
' "$unit_id" "$tx" "$ty"
  fi
}

assert_skill_snapshot() {
  local step_body="$1"
  local enemy_id="$2"
  local hp_before="$3"
  if command -v jq >/dev/null 2>&1; then
    local snap skill_id hp_after
    snap=$(echo "$step_body" | jq -c '.view_snapshot_json // {}')
    skill_id=$(echo "$snap" | jq -r '.lastSkillCast.skillId // empty')
    if [[ -z "$skill_id" ]]; then
      echo "skill smoke: view_snapshot_json missing lastSkillCast after step-lockstep" >&2
      echo "$step_body" >&2
      exit 1
    fi
    hp_after=$(echo "$snap" | jq -r --argjson eid "$enemy_id" '
      (.units // []) | map(select(.id == $eid)) | .[0].hp // empty
    ')
    if [[ -n "$hp_after" && -n "$hp_before" && "$hp_after" -ge "$hp_before" ]]; then
      echo "skill smoke: expected enemy hp drop (before=${hp_before} after=${hp_after})" >&2
      exit 1
    fi
    echo "ok: lastSkillCast.skillId=${skill_id} enemy_hp=${hp_after:-unknown}"
  else
    echo "$step_body" | python3 -c '
import json, sys
enemy_id = int(sys.argv[1])
hp_before = sys.argv[2]
data = json.load(sys.stdin)
snap = data.get("view_snapshot_json") or {}
cast = snap.get("lastSkillCast")
if not cast:
    raise SystemExit("skill smoke: missing lastSkillCast")
hp_after = None
for u in snap.get("units") or []:
    if int(u.get("id")) == enemy_id:
        hp_after = u.get("hp")
        break
if hp_before.isdigit() and hp_after is not None and int(hp_after) >= int(hp_before):
    raise SystemExit(f"skill smoke: expected hp drop {hp_before} -> {hp_after}")
print(f"ok: lastSkillCast.skillId={cast.get('skillId')} enemy_hp={hp_after}")
' "$enemy_id" "$hp_before"
  fi
}

pick_enemy_id_and_hp() {
  local snapshot_json="$1"
  if command -v jq >/dev/null 2>&1; then
    echo "$snapshot_json" | jq -r --argjson owner "$PLAYER_ID" '
      (.units // []) | map(select(.owner != $owner)) | .[0] | "\(.id) \(.hp)"
    '
  else
    echo "$snapshot_json" | python3 -c "
import json, sys, os
owner = int(os.environ.get('PLAYER_ID', '0'))
data = json.load(sys.stdin)
for u in data.get('units') or []:
    if int(u.get('owner', -1)) != owner:
        print(f\"{u['id']} {u.get('hp', '')}\")
        break
"
  fi
}

if [[ -z "${ACCESS_TOKEN:-}" ]]; then
  ACCESS_TOKEN=$(mint_access_token || true)
  if [[ -z "$ACCESS_TOKEN" ]]; then
    echo "error: ACCESS_TOKEN is required (set ACCESS_TOKEN, enable JANUS_HTTP_DEV_MINT + POST /v1/lares/login, or Lares gRPC Login; see docs/janus-http-mirror.md)" >&2
    exit 1
  fi
  echo "==> minted ACCESS_TOKEN via POST ${BASE}/v1/lares/login (smoke/smoke)"
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

if [[ "$SKILL_MODE" -eq 1 ]]; then
  if ! command -v python3 >/dev/null 2>&1; then
    echo "error: KindSkill smoke requires python3 (bridge planning)" >&2
    exit 1
  fi
  read -r SKILL_TO_X SKILL_TO_Y BRIDGE_X BRIDGE_Y <<< "$(plan_skill_bridge "$SNAPSHOT_JSON" "$UNIT_ID")"
  read -r ENEMY_ID ENEMY_HP_BEFORE <<< "$(pick_enemy_id_and_hp "$SNAPSHOT_JSON")"
  echo "==> KindSkill plan: target=(${SKILL_TO_X},${SKILL_TO_Y}) bridge=(${BRIDGE_X},${BRIDGE_Y}) unit_id=${UNIT_ID}"

  bridge_moves=0
  while true; do
    dist=$(unit_cheb_to_target "$SNAPSHOT_JSON" "$UNIT_ID" "$SKILL_TO_X" "$SKILL_TO_Y")
    if [[ "$dist" == "1" ]]; then
      break
    fi
    if [[ "$bridge_moves" -ge "$MAX_BRIDGE_MOVES" ]]; then
      echo "skill smoke: exceeded MAX_BRIDGE_MOVES=${MAX_BRIDGE_MOVES} (still dist=${dist} from target)" >&2
      exit 1
    fi
    read -r move_x move_y <<< "$(next_bridge_move_dest "$SNAPSHOT_JSON" "$UNIT_ID" "$BRIDGE_X" "$BRIDGE_Y")"
    echo "==> bridge move ${bridge_moves}: POST command kind=1 → (${move_x},${move_y})"
    MOVE_BODY=$(post_command 1 "$UNIT_ID" "$move_x" "$move_y")
    assert_command_accepted "bridge move" "$MOVE_BODY"
    STEP_BODY=$(step_lockstep "$LOCKSTEP_STEPS")
    if command -v jq >/dev/null 2>&1; then
      SNAPSHOT_JSON=$(echo "$STEP_BODY" | jq -c '.view_snapshot_json // {}')
    else
      SNAPSHOT_JSON=$(echo "$STEP_BODY" | python3 -c 'import json,sys; print(json.dumps(json.load(sys.stdin).get("view_snapshot_json") or {}))')
    fi
    bridge_moves=$((bridge_moves + 1))
  done

  echo "==> POST ${BASE}/v1/tactical/command (kind=5 skill_id=${SKILL_ID} → ${SKILL_TO_X},${SKILL_TO_Y})"
  CMD_BODY=$(post_command 5 "$UNIT_ID" "$SKILL_TO_X" "$SKILL_TO_Y" "$SKILL_ID")
  assert_command_accepted "KindSkill" "$CMD_BODY"
  STEP_BODY=$(step_lockstep "$LOCKSTEP_STEPS")
  assert_skill_snapshot "$STEP_BODY" "$ENEMY_ID" "$ENEMY_HP_BEFORE"
  echo "ok: battle_id=${BATTLE_ID} unit_id=${UNIT_ID} skill_id=${SKILL_ID} accepted=true bridge_moves=${bridge_moves}"
  exit 0
fi

echo "==> POST ${BASE}/v1/tactical/command (unit_id=${UNIT_ID}, kind=${MOVE_KIND})"
CMD_BODY=$(post_command "$MOVE_KIND" "$UNIT_ID" "$TO_X" "$TO_Y")
assert_command_accepted "command" "$CMD_BODY"

echo "ok: battle_id=${BATTLE_ID} unit_id=${UNIT_ID} accepted=true"
