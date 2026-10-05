# Timeout & hedging — bound the wait, cut the tail

## 0. Why timeout

- One slow dependency parks your goroutines and connections;
  enough slow replies and every worker waits — the stall cascades
  upward while every server looks "busy but fine".
- Timeout turns unbounded waiting into fast failure: the caller
  gets an answer (or a fallback) in bounded time, and capacity
  returns to serve requests that can succeed.

## 1. Tradeoff

|                             |                                                                                        |
| --------------------------- | -------------------------------------------------------------------------------------- |
| ✅ Fail fast                | bounded wait — a dead downstream costs milliseconds, not workers                       |
| ✅ Frees capacity           | timed-out workers serve the next request instead of the lost one                       |
| ✅ Forces fallback          | a deadline is what triggers the degraded path (stale cache, default)                   |
| 🪙 Too tight kills healthy  | a 300ms timeout on a 400ms-p99 dependency amputates good traffic                       |
| 🪙 Client-only timeout lies | the caller is free but the server still burns the full work (abandoned)                |
| 🪙 One timeout for all      | every dependency has its own latency profile — a global value is wrong for all of them |

## 2. Budget rules

| Rule                          | Why                                                               |
| ----------------------------- | ----------------------------------------------------------------- |
| per-dependency timeout        | each downstream gets its own budget from its own p99              |
| client timeout < upstream SLA | the inner call must die before the outer promise breaks           |
| propagate `ctx` end to end    | deadline travels with the request; every layer respects `Done`    |
| server stops on cancel        | quit before writing — the spent work is lost, new work is refused |

## 3. Hedging the tail

|                             |                                                          |
| --------------------------- | -------------------------------------------------------- |
| ✅ Second flight at ~p95    | cuts the tail without doubling the median load           |
| ✅ First success wins | the loser always dies by cancel |
| ✅ Cancel the loser, not the first | fast primary = no hedge fired; winning primary = hedge cancelled |
| 🪙 Hedge POST               | a second write may double-charge — idempotent reads only |
| 🪙 Hedge a saturated server | extra load deepens the hole it was meant to escape       |

## 4. Cause → consequence

| ❌ Cause → consequence                              | ✅ Instead                       |
| --------------------------------------------------- | -------------------------------- |
| no timeout → one slow dep parks all workers         | per-dependency deadline          |
| too-tight timeout → false errors on healthy traffic | budget from p99 + headroom       |
| no propagation → server burns for a gone client     | `ctx` end to end, stop on `Done` |
| hedge without cancel → permanent 2x load            | cancel the loser, always         |
| hedge saturated downstream → deeper hole | hedge only with headroom |

## 5. Option: stale on timeout

Serve last-good (in-memory copy, age-bounded) on `DeadlineExceeded`
instead of erroring: `X-Stale: true` + `stale_served_total` + bound
N minutes past which stale is worse than error. Niche, not default
— worth it only when origin is remote, sometimes sick, and stale
beats error (feeds, catalog, DNS, CDN edge: RFC 5861
`stale-while-revalidate` / `stale-if-error`). Ordinary CRUD:
timeout + error + retry wins by simplicity.
