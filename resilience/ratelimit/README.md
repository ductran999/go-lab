# Ratelimit lab — client-side bucket pacing my own calls

Server on `:8126`. Plain `/work?ms=N` (counts hits) — no bucket
here; throttling lives in the client (this pillar is client-side).
`/stats` shows hits: smooth ~10/sec instead of one spike.

## Run

```bash
make run-server  # plain downstream on :8126
make run-client  # 30 calls paced at 10/sec (ARGS="-n 100 -rate 5" to play)
```

## Try

```bash
# 30 calls at 10/sec burst 10: ~2s wall, server sees smooth traffic
go run ./cmd/client
# ratelimit n=30 rate=10 burst=10 ok=30 shed=0 wall=~2s
# server: {"hits":30}
```

## Docs

- `docs/01-ratelimit.md` — token bucket, fair share, cause
