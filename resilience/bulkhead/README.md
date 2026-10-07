# Bulkhead lab — separate lanes per downstream

Server on `:8125`. `/slow?ms=N` burns N ms, `/fast` answers at
once, `/stats` counts per-endpoint hits.

## Run

```bash
make run-server  # two downstreams on :8125
make run-client  # split vs shared lanes (ARGS="-split=false" to compare)
```

## Try

````bash
# 1. Split lanes: slow saturates its pool, fast stays milliseconds
go run ./cmd/client -split=true
# bulkhead split=true slowest-fast=1ms ...

# 2. Shared lane: fast queues behind slow (~700ms: arrives at
# 300ms, first wave frees at 1000ms)
go run ./cmd/client -split=false
# bulkhead split=false slowest-fast=707ms ...

## Swimlanes (width = time, S = 1s slow, f = 5ms fast)

```text
split=true (slow:2 + fast:2)
┌─ slow lane ─────────────────────────────────┐
│ ████████ ███████ │ ████████████████ │ ██... │ ≈ 3s
│   S1 S2          │   S3 S4          │ S5 S6 │
├─ fast lane ─────────────────────────────────┤
│ ▏▏▏▏▏▏ f1..f6 → done in ms                  │
└─────────────────────────────────────────────┘

split=false (shared:4)
┌─ shared ────────────────────────────────────┐
│ ████████ S1..S4 │ ███ S5 S6 ▏▏ f1 f2 │ f3.. │ fast queues
└─────────────────────────────────────────────┘
→ slowest-fast: 1ms vs ~700ms — same load, only lanes differ
````

```

## Docs

- `docs/01-bulkhead.md` — compartments, tradeoff, cause
```
