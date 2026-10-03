# cron — scheduled work with timeouts

```
scheduler/  ticks + timeout + drain (dumb core)
jobs/       one func per usecase (pure logic)
guards/     overlap/retry/idempotency wrappers (compose outside-in)
cmd/cron/   wiring: scheduler.New(guards.SkipOverlap(jobs.Demo))
```

robfig/cron worker: `@every 5s` demo job (5s work, 2s ctx timeout —
watch it get cancelled), graceful stop on SIGINT/SIGTERM.

```bash
# Change DIR
$ cd jobs/cron
```

## Run

```bash
make run  # watch trigger → deadline → next round; Ctrl-C drains
make escape  # prove `running` escapes to heap (closure outlives setup)
```

## SLI/SLO/SLA (the trio)

- **SLI** (Indicator): the number — uptime 99.9%, p99 < 200ms.
- **SLO** (Objective): internal target — keep SLI above 99.9%.
- **SLA** (Agreement): contract with teeth — miss SLO, pay up.
- SLI measures → SLO promises in-house → SLA signs with customers.

## Why jitter (thundering herd)

100 jobs (or 100 replicas) on `0 2 * * *` fire the same second:
DB/cache stampede, everything slow, alerts fire. Random 0..N delay
per trigger spreads starts — herd becomes a line. Compose with
overlap-drop for stragglers: `SkipOverlap(Jitter(job, 2s))`.
Neither alone suffices (jitter still collides sometimes; overlap
without jitter stampedes together every round).

## Model: one heap cell, many stacks

```text
main frame:      running [heap] ◄──┐ (outlives SkipOverlap return)
                                    │
tick 1 goroutine → callbackFunc ────┘── CAS(false→true) wins → job()
tick 2 goroutine → callbackFunc ────── CAS loses → skip
```

Two goroutines, two stacks, one shared heap cell; CAS picks a single
winner atomically. Setup once (`SkipOverlap`), run N times (ticks).

## Rules

- Every job takes `ctx` and honors `Done()` (timeout/cancel must
  actually stop work, or timeouts are decoration).
- Overruns: timeout < interval means overlapping runs — use
  `cron.SkipIfStillRunning` or singleflight for real jobs.
