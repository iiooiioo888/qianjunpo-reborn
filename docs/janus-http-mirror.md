# Janus HTTP tactical mirrors (dev / Cocos browser preview)

Janus `:8090` exposes JSON mirrors of gRPC battle APIs for clients without gRPC (Cocos Live preview).

| Method | Path | gRPC equivalent |
|--------|------|-----------------|
| `GET` | `/v1/tactical/snapshot?battle_id=…` | `GetBattleSnapshot` |
| `POST` | `/v1/tactical/command` | `SubmitTacticalCommand` |

## POST `/v1/tactical/command`

Request body (snake_case, aligned with `TacticalCommandClient.ts`):

```json
{
  "session_id": "optional",
  "battle_id": "default/0",
  "player_id": 0,
  "kind": 1,
  "unit_id": 1,
  "to_x": 5,
  "to_y": 8
}
```

Response:

```json
{
  "accepted": true,
  "reject_reason": "",
  "lockstep_frame": 0,
  "state_hash": 0
}
```

### Environment

| Variable | Default | Meaning |
|----------|---------|---------|
| `JANUS_HTTP_TACTICAL_COMMAND` | enabled (`1`) | Set to `0` / `false` / `disabled` to return HTTP 503 with `accepted: false` instead of forwarding to Roma. |

### Smoke test (with `make compose-up`)

```bash
curl -sS -X POST 'http://127.0.0.1:8090/v1/tactical/command' \
  -H 'Content-Type: application/json' \
  -d '{"battle_id":"default/0","player_id":0,"kind":1,"unit_id":1,"to_x":5,"to_y":8}'
```
