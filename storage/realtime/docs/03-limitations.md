# Limitations

**TL;DR:** 8KB payloads, transient delivery, one backend per listener, no transaction-pooler support.

Keywords: payload cap, pooling, replay gap.

## NOTIFY limits

- **8000-byte payload cap.** Big rows fail the trigger. Send `{id, tenant_id, kind}`, let clients fetch full rows.
- **Transient.** No listener = lost. No replay, no offsets, at-most-once.
- **Commit-bound.** Fires after COMMIT only. Good (no phantom events), but no pre-commit hooks either.

## LISTEN + pooling

- One backend per listener. Never a pooled connection (this server uses a dedicated `pgx.Connect` outside the dbkit pool).
- PgBouncer **transaction mode breaks LISTEN**. Session pooling or direct connections only.
- One connection, many channels. Scale out with one notifier + internal bus (Redis/NATS), not one LISTEN conn per instance.

## Stream gaps (this PoC)

- Tail-live only: late connecters miss history. Production shape = replay recent N rows, then tail (plus `Last-Event-ID` resume).
- No auth on `/stream` yet: tenant rides a query param. Next step is JWT middleware (same pattern as the rls lab).
