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

Ties: `../network/headers/cache/` (HTTP caching headers over these
backends), `../observability/` (trace the queries).

## Roadmap

- [x] `pgbouncer/` — pooling demo: transaction mode, gotchas, pgbench compare
- [ ] `gorm-vs-pgx.md` — compare: ergonomics vs control, hooks vs raw, pool behavior (see `gorm/docs/02-gorm-vs-pgx.md`)
- [x] `pgbadger/` — log analysis: slow queries, forensic loop, detection chain
- [x] `cassandra/` — wide-column: timeline schema, QUORUM write + ONE read demo
- [ ] `elasticsearch/` — full-text search: analyzers, relevance scoring
