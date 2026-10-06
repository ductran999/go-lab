# Breaker lab — fail fast, probe back

Server on `:8124`. `/flaky` fails the first `?fail=N` hits (then
200), `/reset?fail=N` rearms, `/stats` counts hits.

```
client call
  ├─▶ breaker [closed]    ──▶ downstream (failures counted)
  ├─▶ breaker [open]      ──▶ ErrOpen, downstream untouched
  └─▶ breaker [half-open] ──▶ 1 probe ──▶ downstream
                                         ├─ ok   → closed
                                         └─ fail → open (cooldown restarts)
```

## Run

```bash
make run-server  # killable downstream on :8124
make run-client  # dead 8s (opens) → healed (probe closes)
make run-gobreaker  # same story, borrowed wheel
```

## Try

```bash
# hands-free lifecycle: dead → open → healed → closed
go run ./cmd/client
# state: closed -> open
# rejected (state open)      <- fast, zero downstream load
# ...
# --- phase 2: downstream healed, waiting for the probe
# state: open -> half-open
# state: half-open -> closed
# final: state=closed calls=.. rejected=.. probes=1
curl -s localhost:8124/stats  # hits froze while open

# borrowed wheel: same lifecycle on sony/gobreaker — outputs match,
# except OnStateChange also prints half-open (stdlib only implies it)
go run ./cmd/gobreaker
# state: closed -> open
# state: open -> half-open
# state: half-open -> closed
```

## Docs

- `docs/01-circuit-breaker.md` — states, tradeoff, retry/timeout play, cause
