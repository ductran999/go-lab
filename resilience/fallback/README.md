# Fallback lab — degrade on purpose, inside bounds

Server on `:8127`. `/flaky` fails the first `?fail=N` hits (then
200 with a rising price, so fresh vs stale is visible),
`/reset?fail=N` rearms, `/stats` counts hits.

```text
request
 └─▶ primary (300ms deadline)
      ├─ ok ──▶ fresh (stores last-good)
      └─ fail ──▶ last-good age < bound? ── yes ──▶ stale
                  │                              (X-Source: stale)
                  └─ no ──▶ strict? ── no ──▶ default
                             │                 (X-Source: default)
                             └─ yes ──▶ error (honest, core paths)
```

## Run

```bash
make up         # shared redis on :6379 (L2 copy)
make run-server  # killable primary on :8127
make run-client  # ladder walk (ARGS="-n 3 -strict" to compare)
```

## Try (one process per block — L1 lives and dies with it)

```bash
# 1. L1 ladder in one run: primary dies after 2 hits (L2 off)
curl -s 'localhost:8127/reset?dieafter=2'
go run ./cmd/client -n 5 -redis ""
# fresh: {"price":101}
# fresh: {"price":102}
# stale: {"price":102}   <- L1, primary dead mid-run
# stale: {"price":102}
# stale: {"price":102}
# sources: map[fresh:2 stale:3]

# 2. Bound expiry: seed, age past 1s, then fail → default
curl -s 'localhost:8127/reset?dieafter=1'
go run ./cmd/client -n 2 -redis "" -stale 1s -gap 1500ms
# fresh: {"price":101}
# default: unknown        <- 1.5s old copy refused

# 3. Strict (core paths): same setup, honest error instead
curl -s 'localhost:8127/reset?dieafter=1'
go run ./cmd/client -n 2 -redis "" -strict
# fresh: {"price":101}
# error: get http://localhost:8127/flaky: status 500

# 4. Shared L2 (needs make up): process A seeds, fresh process B
# has empty L1 yet serves stale-redis while primary stays dead
curl -s 'localhost:8127/reset?fail=0'
go run ./cmd/client -n 1                          # fresh, seeds L1+L2
curl -s 'localhost:8127/reset?fail=100'
go run ./cmd/client -n 1                          # new process:
# stale-redis: {"price":101}   <- L1 empty, L2 hit
```

## Docs

- `docs/01-fallback.md` — ladder, scope, cause
