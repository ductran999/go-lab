# Fallback — degrade on purpose, inside bounds

Every layer above can still say no: attempts out, circuit open,
deadline gone. Fallback is the planned answer instead of an error:
stale, default, or queued — each with a bound past which error is
more honest than data.

## Ladder (in order, first hit wins)

| Step          | Serves                   | Bound                                            |
| ------------- | ------------------------ | ------------------------------------------------ |
| primary       | fresh truth              | timeout + retry + breaker above                  |
| stale         | last-good copy           | age < N (past N, stale lies — refuse it)         |
| default       | static safe value        | only where "unknown" harms less than error       |
| error / queue | honest failure, or defer | core paths (money): fail fast + idempotent retry |

## Scope: edge, never core

Edge = paths where wrong-temporary beats error (feed, config,
catalog display, recommendations). Core = paths where wrong
beats nothing (money, inventory, votes — strict mode only).

|                                    |                                                     |
| ---------------------------------- | --------------------------------------------------- |
| ✅ Reads, recomputable, deferrable | feed, config, outbox-queued writes                  |
| ❌ Strong consistency              | balance, inventory, votes — wrong beats error never |

Fallback trades correctness for availability (CAP's A, declared).
Every fallback names its bound: staleness N minutes, queue drain
X, default labeled `X-Source` so dashboards tell fresh from stale.

## Tradeoff

|                   |                                                                        |
| ----------------- | ---------------------------------------------------------------------- |
| ✅ Availability   | user gets an answer, not an error — incident without pager storm       |
| ✅ Buys time      | degraded service breathes while the primary recovers                   |
| ✅ Cheap          | a map, a bound, a header — no infra                                    |
| 🪙 Correctness    | stale or default served as truth — never on core                       |
| 🪙 RAM            | one copy per key — bound count (LRU) + age (TTL), hot keys only        |
| 🪙 Divergence     | fresh here, stale there across instances until all refresh             |
| 🪙 Hidden rot     | success-after-stale paints dashboards green — count sources separately |
| 🪙 Refresh burden | no refresh path means stale forever, not stale temporarily             |

## In the wild

| Who                 | Fallback                                       | Bound/label                                 |
| ------------------- | ---------------------------------------------- | ------------------------------------------- |
| Netflix (Hystrix)   | recommendations down → static top-popular list | default + owning team per fallback          |
| Cloudflare/Akamai   | dead origin → over-TTL cached copy             | RFC 5861 `stale-if-error`                   |
| DNS resolvers       | upstream fail → serve-stale                    | capped past-TTL + flag                      |
| stock tickers       | 15-min delayed quotes                          | stale-by-design, labeled "delayed"          |
| flight/hotel search | cached price + "prices may change"             | disclaimer as the bound                     |
| mobile offline      | last synced data                               | offline banner, refresh on reconnect        |
| payment queues      | accept now, outbox-process later               | defer (slow, never wrong) + idempotency key |

Common thread: every one labels the bound publicly — silent
fallback is the bug, declared fallback is the feature.

## Cause → consequence

| ❌ Cause → consequence                                      | ✅ Instead                                                                               |
| ----------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| fallback on core → wrong money                              | strict mode: fail fast, user retries                                                     |
| unbounded stale → yesterday's price                         | age bound, refuse past N                                                                 |
| unbounded copies → RAM grows per key                        | bound count (LRU) + age (TTL); hot keys only                                             |
| per-instance L1 → same request gets fresh here, stale there | shared L2 narrows it; readiness only after first seed; accept temporary skew on the edge |
| silent fallback → green lies                                | `X-Source` header + `fallback_served_total` counter                                      |
| no refresh path → stale forever                             | background refresh or next-try-primary                                                   |
