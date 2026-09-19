# Compare to CDC

**TL;DR:** NOTIFY = app-level, transient, zero ops. CDC = infra-level, durable, replayable. Ephemeral UI vs business backbone.

Keywords: at-most-once, at-least-once, outbox.

|                | LISTEN/NOTIFY (this PoC)                 | CDC (Debezium)                               |
| -------------- | ---------------------------------------- | -------------------------------------------- |
| Level          | Application (trigger SQL)                | Infrastructure (WAL reader)                  |
| Payload        | Chosen explicitly, max **8000 bytes**    | All changes, before/after images             |
| Durability     | Transient: no listener = lost, no replay | Persistent log: replay + order               |
| Delivery       | At-most-once                             | At-least-once (+ idempotent consumers)       |
| Touch app code | Yes (trigger per event)                  | No (reads WAL independently)                 |
| Ops            | Zero (Postgres built-in)                 | Kafka Connect cluster                        |
| Middle ground  | —                                        | Logical decoding (`pgoutput` slot, no Kafka) |

## Scope rule

- **Ephemeral** (live feeds, presence, progress): NOTIFY/SSE. A lost message costs a UI line.
- **Business** (order → payment → inventory sagas): durable log. NOTIFY-built sagas demo well and fail in production (one drop = stuck saga, no recovery).
- **Both**: transactional outbox — business row + event row in one transaction (atomic), relayed to a durable log. DB atomicity, log durability.
