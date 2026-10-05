# Timeout lab — deadline propagation and hedging

Server on `:8121`. `/slow?ms=N` burns N ms (quits early on client
cancel), `/fast` answers at once, `/stats` counts started vs
finished — the gap is work abandoned by timeouts.

## Run

```bash
make run-server  # downstream on :8121
make run-client  # three calls: none, deadline, hedged
```

## Try

```bash
# 1. No deadline pays the full sleep
time curl -s 'localhost:8121/slow?ms=2000'
# ~2s

# 2. Client timeout fails fast; server still burned 300ms of work
time curl -s --max-time 0.3 'localhost:8121/slow?ms=2000'
# ~0.3s, curl exit 28
curl -s localhost:8121/stats
# {"started":2,"finished":1,"abandoned":1}

# 3. Hedged client (300ms trigger) finishes with the fast flight
go run ./cmd/client
# 1. no deadline: took ~2s
# 2. 300ms deadline: took ~300ms err=context deadline exceeded
# 3. hedged after 300ms: took ~300ms + body {"ok":true}
```

## Docs

- `docs/01-timeout-hedging.md` — why, tradeoff, budgets, hedging, cause

## Slowloris proof (server + drip client, no test fakes)

```bash
# terminal 1: guarded server, 800ms header budget
go run ./cmd/server -header-ms 800
# terminal 2: drip one header line per 200ms, never finishing
go run ./cmd/drip
# server cut after 4 lines / ~800ms: write: connection reset by peer

# unguarded: default 5s budget outlasts a short drip
go run ./cmd/drip -lines 10
# server still waiting after 10 lines / ~2s
```
