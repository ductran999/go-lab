# Realtime PoC

Postgres `LISTEN/NOTIFY` bridged to browser SSE: DB row in, live event
out. Joins the `sse` lab (fan-out) with the `storage` labs (Postgres
as source of truth).

Deep-dive notes in [`docs/`](docs/): pipeline, CDC compare, limitations.

```bash
# Change DIR
$ cd storage/realtime
```

## Setup

```bash
cp .env.example .env   # edit if needed
make up                # postgres:5434 + migrate up
make run-server        # API + SSE on :8082, frontend at /
```

- Postgres: `localhost:5434`. Channel: `events`, payload = inserted row JSON.
- Table: `app.events (tenant_id, kind, payload, created_at)` + seed rows.

## Try it

```bash
# terminal 1: stream tenant 1 (blocks, prints data: lines)
curl -N 'http://localhost:8082/stream?tenant_id=1'
# terminal 2: write (trigger NOTIFYs, terminal 1 prints it)
curl -s -X POST 'http://localhost:8082/events' \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":1,"kind":"ping","payload":{"msg":"hello"}}' | jq
# or open http://localhost:8082/ and click through
```

## Trigger check (psql only, no code)

```sql
-- session A: listen
LISTEN events;
-- session B: insert
INSERT INTO app.events (tenant_id, kind, payload)
VALUES (1, 'ping', '{"msg":"hello"}');
-- session A receives: Asynchronous notification "events" with payload ...
```

## Cleanup

```bash
make down    # stop, keep data
make clean   # stop + remove postgres_realtime_data volume
```

## Next

- [x] DB + trigger, Go listener, SSE + writer, frontend
- [ ] Stream replay (recent N rows, then live tail + `Last-Event-ID`)
- [ ] JWT tenant filter on `/stream` (same pattern as the rls lab)
