# Timeout — bound the wait

One slow dependency parks goroutines and connections until every
worker waits. Timeout turns unbounded waiting into fast failure —
bounded wait, freed capacity, and a trigger for the fallback path.

## Tradeoff

|                             |                                                                                        |
| --------------------------- | -------------------------------------------------------------------------------------- |
| ✅ Fail fast                | a dead downstream costs milliseconds, not workers                                      |
| ✅ Frees capacity           | timed-out workers serve the next request instead of the lost one                       |
| ✅ Forces fallback          | a deadline is what triggers the degraded path                                          |
| 🪙 Too tight kills healthy  | a 300ms timeout on a 400ms-p99 dependency amputates good traffic                       |
| 🪙 Client-only timeout lies | the caller is free but the server still burns the full work (abandoned)                |
| 🪙 One timeout for all      | every dependency has its own latency profile — a global value is wrong for all of them |

## Budgets

| Rule                          | Why                                                               |
| ----------------------------- | ----------------------------------------------------------------- |
| per-dependency timeout        | each downstream gets its own budget from its own p99 (`p99 × 2–3`) |
| client timeout < upstream SLA | the inner call must die before the outer promise breaks           |
| propagate `ctx` end to end    | deadline travels with the request; every layer respects `Done`    |
| server stops on cancel        | quit before writing — the spent work is lost, new work is refused |

## Deadline vs cancel

Same `ctx.Done()` channel, opposite diseases — count separately.

| Signal                        | Means                                                                                      | Look at                                       |
| ----------------------------- | ------------------------------------------------------------------------------------------ | --------------------------------------------- |
| `DeadlineExceeded` high       | downstream too slow, or budget too tight                                                   | p99 vs timeout, downstream health             |
| `Canceled` high, deadline low | callers abandoning: disconnects, hedge losers, deploys, superseded requests (autocomplete) | client behavior, deploy frequency, hedge rate |
| both spike together           | the whole path is stuck — upstream deadline firing while callers flee                      | upstream SLA first, then downstream           |

## Cause → consequence

| ❌ Cause → consequence                              | ✅ Instead                       |
| --------------------------------------------------- | -------------------------------- |
| no timeout → one slow dep parks all workers         | per-dependency deadline          |
| too-tight timeout → false errors on healthy traffic | budget from p99 + headroom       |
| no propagation → server burns for a gone client     | `ctx` end to end, stop on `Done` |
