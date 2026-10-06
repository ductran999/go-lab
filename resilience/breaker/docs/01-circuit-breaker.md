# Circuit breaker — stop hammering the dead

Retry assumes recovery is near; when the downstream is truly
down, retries become a self-DDoS with a success story attached.
The breaker watches outcomes and opens the circuit: calls fail
fast without touching the wire until a probe proves recovery.

## States

| State     | Does                            | Enters when                                            |
| --------- | ------------------------------- | ------------------------------------------------------ |
| closed    | traffic flows, failures counted | half-open probe succeeds                               |
| open      | fail fast, zero downstream load | failure threshold crossed (e.g. 5 fails, or 50% of 20) |
| half-open | one probe flight allowed        | open timeout elapses (e.g. 30s)                        |

Probe fails → back to open (timeout restarts). Probe succeeds →
closed, counters reset. One probe at a time: a herd of probes is
a stampede with a permission slip.

## Tradeoff

|                    |                                                                                                       |
| ------------------ | ----------------------------------------------------------------------------------------------------- |
| ✅ Stops the bleed | dead downstream gets zero load — room to recover                                                      |
| ✅ Fails fast      | callers get instant error → fallback path triggers now, not after timeout × attempts                  |
| ✅ Auto recovery   | half-open probe closes the loop without human intervention                                            |
| 🪙 Flapping        | threshold too tight on a flaky (not dead) downstream → circuit oscillates, good traffic rejected      |
| 🪙 Probe cost      | the single probe that fails still pays full timeout                                                   |
| 🪙 State scope     | per-instance breaker sees partial truth (3 instances, 3 opinions) — shared state or accept divergence |

## Plays with retry/timeout

| Rule                                 | Why                                                                   |
| ------------------------------------ | --------------------------------------------------------------------- |
| timeout inside, breaker outside      | per-try deadline bounds each attempt; breaker bounds all attempts     |
| breaker counts retryable only        | 403/400 must not trip the breaker — dead errors are not outage signal |
| open circuit short-circuits retry    | no point retrying 5 times into an open breaker — fail fast once       |
| half-open probe gets its own timeout | a probe hanging forever holds recovery hostage                        |

Nesting order (in → out):

```go
try := func() error { // per-try deadline
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)

	defer cancel()

	return fetch.Do(ctx, url)
}
attempt := func() error { // bounded retry, retryable only
	return retry.Call(ctx, 5, 50*time.Millisecond, true, retry.RetryableStatus, try)
}
err := breaker.Call(attempt) // dead downstream: fail fast, zero attempts
```

## Cause → consequence

| ❌ Cause → consequence                     | ✅ Instead           |
| ------------------------------------------ | -------------------- |
| no breaker → retries DDoS the dead         | open on threshold    |
| threshold on all errors → flapping on 403s | count retryable only |
| herd of probes → mini-stampede on recovery | one probe at a time  |
| no auto-close → human resets at 3am        | half-open probe loop |
