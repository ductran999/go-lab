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

| ❌ Cause → consequence                   | ✅ Instead             |
| ---------------------------------------- | ---------------------- |
| hedge without cancel → permanent 2x load  | cancel the loser, always |
| hedge saturated downstream → deeper hole | hedge only with headroom |

## Option: stale on timeout

Serve last-good (in-memory copy, age-bounded) on `DeadlineExceeded`
instead of erroring: `X-Stale: true` + `stale_served_total` + bound
N minutes past which stale is worse than error. Niche, not default
— worth it only when origin is remote, sometimes sick, and stale
beats error (feeds, catalog, DNS, CDN edge: RFC 5861
`stale-while-revalidate` / `stale-if-error`). Ordinary CRUD:
timeout + error + retry wins by simplicity.
