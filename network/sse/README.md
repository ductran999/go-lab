# SSE lab — server pushes over plain HTTP

Companion to `../websocket/`: same one-way need, opposite tool.
Web page on `:8095`, API on `:8096`.

## Run

```bash
make run-server  # API on :8096
make run-web     # page on :8095
```

## Try

```bash
# Headers first: plain HTTP response, no upgrade
curl -s -D- -o /dev/null --max-time 3 \
  -H 'Origin: http://localhost:8095' localhost:8096/events

# 3 events then hang (ticker 1s)
curl -sN --max-time 3.5 localhost:8096/events

# Kill after 2, then resume where it stopped
curl -sN --max-time 2.5 'localhost:8096/events?kill_after=2'
curl -sN --max-time 2.5 -H 'Last-Event-ID: 2' localhost:8096/events
# ids continue 3,4,... nothing repeats, nothing skipped
```

Browser on `:8095` → Connect, watch `id=` climb. "kill after 3" →
error line → auto-retry resumes at 4 (see `Last-Event-ID` in DevTools).

## Docs

- `docs/01-sse-basics.md` — wire format, headers, resume
- `docs/02-ws-vs-sse.md` — when to pick which
