# Resilience — fail fast, retry smart, degrade on purpose

> TL;DR: timeouts bound the wait, retries survive the blip,
> breakers stop the bleed, singleflight kills the stampede.
> Order of survival: timeout → retry+jitter → breaker →
> singleflight → bulkhead/ratelimit → fallback.

| Lab | Solves | Port |
|---|---|---|
| `retry/` | backoff + jitter, retry budgets | `:8122` |

Rule of the pillar: every lab runs a flaky downstream next to
the client, counts what happened (`/stats`), and proves the
technique with numbers — not adjectives.
