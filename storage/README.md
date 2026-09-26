# Storage — data lives here

Postgres-centric labs plus cache backends: one pillar for "where
bytes rest" (vs `network/` where they travel).

## Map

| Dir          | What                                                          |
| ------------ | ------------------------------------------------------------- |
| `postgrest/` | REST from Postgres, JWT, multi-tenant RLS                     |
| `rls/`       | Row Level Security: policies, benchmarks, Go integration      |
| `realtime/`  | Postgres NOTIFY → SSE bridge (JWT + resume)                   |
| `cache/`     | Backends compared: Redis, Memcached, Dragonfly, KeyDB, in-mem |

Ties: `../network/http/cache/` (HTTP caching headers over these
backends), `../observability/` (trace the queries).
