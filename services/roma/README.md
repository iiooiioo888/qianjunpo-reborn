# Roma zone service

Authoritative battle state (HP, position, buffs, lockstep frame) lives **only in Roma process memory**.

Redis (and MySQL) may be used for:

- Session routing keys / gateway stickiness
- Non-authoritative caches (static tables, chat history)

Redis must **NOT** cache or persist authoritative combat fields (HP, position, buff state) as source of truth.
