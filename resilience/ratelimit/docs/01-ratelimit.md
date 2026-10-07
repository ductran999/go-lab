# Ratelimit — token bucket at the door

Popularity is a DDoS you asked for: launches, retries, and one
buggy loop can 100× normal traffic in a minute. Ratelimit caps
admissions (requests/second with burst) — excess fails fast with
`429 + Retry-After` instead of queuing into collapse.

## Mechanism

| Piece                  | Does                                                                                        |
| ---------------------- | ------------------------------------------------------------------------------------------- |
| token bucket           | N tokens/sec refill, burst up to M — smooth rate, tolerant spikes                           |
| server-side (inbound)  | protects me: per client/IP/key, 429 the excess                                              |
| client-side (outbound) | protects downstream: throttle my own calls before they stampede it                          |
| headers                | `X-RateLimit-Limit/Remaining/Reset` + `Retry-After` — the client can back off intelligently |

## Tradeoff

|                           |                                                                                          |
| ------------------------- | ---------------------------------------------------------------------------------------- |
| ✅ Collapse prevented     | overload becomes clean 429s, not timeouts and OOMs                                       |
| ✅ Fair sharing           | one greedy client can't starve the rest                                                  |
| ✅ Retry talks to it      | 429 + `Retry-After` is the only error that tells the client exactly how to behave        |
| 🪙 Legit traffic rejected | threshold below real peaks = self-inflicted outage (size from p99 + headroom, per key)   |
| 🪙 Distributed counting   | per-instance buckets let N×limit through — shared counter (Redis) or accept the multiple |

## Cause → consequence

| ❌ Cause → consequence                   | ✅ Instead                               |
| ---------------------------------------- | ---------------------------------------- |
| no limit → spike queues into OOM         | token bucket, fail fast with 429         |
| global limit → one client starves all    | per-key buckets, fair share              |
| silent reject → clients retry blindly    | `Retry-After` + remaining headers        |
| per-instance only → N× the intended rate | shared counter, or size for the multiple |
