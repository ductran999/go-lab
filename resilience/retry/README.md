# Retry lab — backoff, jitter, budgets

Server on `:8122`. `/flaky` fails the first `?fail=N` hits (then
200), `/stats` shows hits in 50ms buckets, `/reset?fail=N` rearms.

## Run

```bash
make run-server  # flaky downstream on :8122
make run-client  # 10 clients x 5 attempts, jitter on
```

## Try

```bash
# 1. Lockstep retries: tall narrow spikes in the buckets
curl -s 'localhost:8122/reset?fail=10'
go run ./cmd/client -n 10 -jitter=false -baseMs 100
# buckets {"0":10,"2":10} — wave 1 together, every retry at +100ms together

# 2. Jittered retries: same success, spread load (and faster wall)
curl -s 'localhost:8122/reset?fail=10'
go run ./cmd/client -n 10 -jitter=true -baseMs 100
# buckets like {"0":17,"1":3} — retries scatter inside [0,100ms]

# 3. Budget: attempts run out before the server heals
curl -s 'localhost:8122/reset?fail=100'
go run ./cmd/client -n 2 -attempts 3
# failed=2 — retries.exhausted, and that's correct: bounded loss

# 4. Dead errors stop at once: 403 is never retried
go run ./cmd/client -n 3 -path deny
# failed=3, wall ~ms, server hits == 3 — one touch each, zero retries
```

## Docs

- `docs/01-retry-jitter.md` — backoff, full jitter, retry rules
