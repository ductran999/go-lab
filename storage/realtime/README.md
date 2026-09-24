# Realtime PoC

Postgres `LISTEN/NOTIFY` bridged to browser SSE: DB row in, live event
out. Joins the `sse` lab (fan-out) with the `storage` labs (Postgres
as source of truth).

Deep-dive notes in [`docs/`](docs/): pipeline, CDC compare, limitations,
auth + resume.

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
TOKEN=$(curl -s -X POST 'http://localhost:8082/token' \
  -H 'Content-Type: application/json' -d '{"tenant_id":1}' | jq -r .token)
# terminal 1: authenticated stream (blocks, prints id: + data: lines)
curl -N "http://localhost:8082/stream?token=$TOKEN"
# terminal 2: write as tenant 1 (tenant from token, not body)
curl -s -X POST 'http://localhost:8082/events' \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"kind":"ping","payload":{"msg":"hello"}}' | jq
# resume: kill terminal 1, write more, reconnect with Last-Event-ID
curl -N "http://localhost:8082/stream?token=$TOKEN" -H 'Last-Event-ID: 3'
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
- [x] Stream replay (missed rows, then live tail + `Last-Event-ID`)
- [x] JWT tenant filter on `/stream` (same pattern as the rls lab)
