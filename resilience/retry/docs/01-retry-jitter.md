# Retry & jitter — survive the blip without stampeding it

## Why retry

- Most failures are transient blips, not verdicts: GC pause,
  rolling deploy, DNS hiccup, split-network noise, a 200ms
  overload. The next try, milliseconds later, succeeds.

## Tradeoff

| Side                    | What happens                                                                                                                                          |
| ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| ✅ Gain: UX             | blips turn into invisible latency — the user waits 100ms more instead of seeing an error for a request that succeeds on retry 1                       |
| ✅ Gain: ops sleeps     | self-healing blips never page — fewer 3am alerts for a 502 that would have passed on attempt 2                                                        |
| ✅ Gain: deploy anytime | rolling restarts cause seconds of 502/503 by design; retry absorbs them, so deploys stop needing maintenance windows                                  |
| ✅ Gain: cheapest lever | dozens of lines, zero infra — vs over-provisioning or failover for the same blip                                                                      |
| 🪙 Price: latency       | every attempt spends the SLO budget — attempts × per-try timeout must fit inside the caller's deadline, or the retry outlives its purpose             |
| 🪙 Price: load          | N clients × M attempts = N×M load on a server that just stumbled — without backoff + jitter + bounded attempts, the cure self-DDoSes                  |
| 🪙 Price: signal        | success-after-retry paints the dashboard green while users waited — count attempts/retries separately, or the outage hides behind a 100% success rate |
| 🪙 Price: bill          | metered downstreams (paid APIs, LLM tokens) charge per attempt — unbounded retry multiplies real money, not just load                                 |
| 🪙 Price: correctness   | a retried non-idempotent POST double-charges, sends twice, deducts twice — not extra cost but wrong results (avoided by idempotency keys)             |
| 🪙 Price: holding       | backoff sleeps hold a goroutine + connection each — a big spike times a long sleep exhausts memory/fds before attempts run out                        |

## Retryable matrix

| Signal                   | ✅/❌                                    |
| ------------------------ | ---------------------------------------- |
| timeout, reset, 429, 503 | ✅ bounded (429 honors `Retry-After`)    |
| 500                      | ✅ transient only, tightly bounded       |
| 400 / 403 / 404 / 422    | ❌ permanent — attempt 100, same verdict |
| ctx done                 | ❌ propagate, never retry                |

## Cause → consequence

| ❌ Cause → consequence             | ✅ Instead         |
| ---------------------------------- | ------------------ |
| no backoff → stampede              | backoff ×2, capped |
| no jitter → second herd            | full jitter        |
| unbounded → infinite load          | max 5 attempts     |
| uncapped sleep → outlives deadline | cap 5s             |
| retry 403/400 → false traffic      | never retry dead   |
| plain POST retry → double-charge   | idempotency key    |
| ignore ctx → work after give-up | propagate ctx |
