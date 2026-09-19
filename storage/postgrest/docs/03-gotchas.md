# Gotchas

**TL;DR:** 403 → check grants. POST fails → check sequences. Connection fails → check password match. Stale API → reload schema cache.

Keywords: privileges, sequences, idempotency, schema cache.

## Security

- **Grants ↔ verbs.** `SELECT`-only → POST/PATCH/DELETE = `401`/`403`. Not a PostgREST bug.
- **PII never in query.** URLs leak into logs/history. Email/tokens → RPC/POST body.
- **Tenant never client-supplied.** `tenant_id=eq.X` in URL = IDOR by design. JWT → RLS (see `02-auth-model.md`).
- **Heavy embeds = DoS surface.** Deep `select=*,rel(*)` joins exhaust the DB. Mitigate: `statement_timeout`, `PGRST_MAX_ROWS`, RLS.

## Performance

- **Complex filters → views/RPC.** Long URLs hit length limits (`414`), fragment cache, unreadable logic. Hide complexity server-side.
- **Schema cache.** DDL change → reload (restart or `NOTIFY pgrst, 'reload schema'`). Stale cache = stale API.

## Operations

- **Identity keys need sequence grants.** No `USAGE, SELECT ON SEQUENCES` → POST fails.
- **Migrator one-shot.** Runs `up`, exits. Version in `schema_migrations`. Applied migration edited → never re-runs. Fix: manual delta or new migration.
- **Naming.** 6-digit prefix = version (`000001_*.up.sql` + `*.down.sql`, same version both files).
