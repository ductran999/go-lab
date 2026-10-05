# Hedging — cut the tail with a second flight

Slow first flight? Fire again at ~p95, first success wins, loser
dies by cancel. Cancel the loser, not the first: fast primary
means no hedge fired; winning primary means the hedge is cancelled.

|                                    |                                                                  |
| ---------------------------------- | ---------------------------------------------------------------- |
| ✅ Second flight at ~p95           | cuts the tail without doubling the median load                   |
| ✅ First success wins              | the loser always dies by cancel                                  |
| ✅ Cancel the loser, not the first | fast primary = no hedge fired; winning primary = hedge cancelled |
| 🪙 Hedge POST                      | a second write may double-charge — idempotent reads only         |
| 🪙 Hedge a saturated server        | extra load deepens the hole it was meant to escape               |

## Cause → consequence

| ❌ Cause → consequence                   | ✅ Instead               |
| ---------------------------------------- | ------------------------ |
| hedge without cancel → permanent 2x load | cancel the loser, always |
| hedge saturated downstream → deeper hole | hedge only with headroom |

## Fastest of N — hedging with all flights at once

Not the tasty default — the last resort. Pecking order for the
tail: right timeout first, hedge-on-tail second, race only when
hedging still isn't fast enough. Worth it solely when p99 SLO
outweighs server cost, replicas idle, AZ-decorrelated, and reads
only; missing any one is over-engineering.

Same request to N replicas simultaneously, first answer wins,
losers cancelled.

|                                      |                                                                                                     |
| ------------------------------------ | --------------------------------------------------------------------------------------------------- |
| ✅ Shortest tail physically possible | no waiting for p95 trigger — the race starts at 0ms                                                 |
| ✅ Survives one slow replica         | straggler loses quietly, user never knows                                                           |
| 🪙 Always N× load                    | every request costs N, even when all replicas fast                                                  |
| 🪙 Correlated slowness wins          | same-AZ outage slows all flights equally — race decorrelated replicas (AZ/rack apart) or don't race |

## Scope: effectively-once only

|                                  |                                                                                     |
| -------------------------------- | ----------------------------------------------------------------------------------- |
| ✅ Safe reads (GET/HEAD/OPTIONS) | no side effects — race freely                                                       |
| ⚠️ Keyed writes                  | idempotency key + server-side dedup required (key dedups the loser, not the cancel) |
| ❌ Plain writes                  | the loser may commit before cancel lands — cancel stops waiting, never undoes       |

## Counting: demand vs load

One race = 1 demand but N flights of load (losers burn real
server work). Tag hedge flights (`X-Hedged: true`, label
`hedged=true` — never a route label) and count both: demand for
traffic truth, load for capacity planning. Budget the race ratio
(only the requests whose p99 matters) instead of racing everything.

## Option: stale on timeout

Serve last-good (in-memory copy, age-bounded) on `DeadlineExceeded`
instead of erroring: `X-Stale: true` + `stale_served_total` + bound
N minutes past which stale is worse than error. Niche, not default
— worth it only when origin is remote, sometimes sick, and stale
beats error (feeds, catalog, DNS, CDN edge: RFC 5861
`stale-while-revalidate` / `stale-if-error`). Ordinary CRUD:
timeout + error + retry wins by simplicity.
