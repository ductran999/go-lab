# Singleflight — one key, one flight, everyone shares

Hot key expires → N concurrent misses → N identical expensive
queries hit the DB at once (stampede). Singleflight collapses
them: first caller executes, rest attach and share the result.

## Mechanism

| Role                            | Does                                                |
| ------------------------------- | --------------------------------------------------- |
| leader (first by key)           | executes the real call                              |
| followers (same key, in-flight) | wait, then share leader's result — success or error |
| key forgotten on done           | next miss after completion starts a fresh flight    |

Use `golang.org/x/sync/singleflight` — map + mutex + `WaitGroup`
is 30 lines to hand-roll and a lifetime to get right under
cancellation. (Same rule as retry: borrow the wheel.)

## Tradeoff

|                          |                                                                             |
| ------------------------ | --------------------------------------------------------------------------- |
| ✅ Kills the stampede    | N misses → 1 query, no extra infra (in-process)                             |
| ✅ Latency collapses too | followers pay one shared wait instead of queueing N queries                 |
| 🪙 Slowest dictates all  | leader's latency becomes everyone's — one straggler holds N callers         |
| 🪙 Errors are shared     | leader fails → all fail; retry must wrap the flight (once), not each caller |
| 🪙 In-process only       | multi-instance needs the distributed twin (Redis lock, LB coalescing)       |
| 🪙 Key must truly match  | per-user/per-request keys never share — and each distinct key holds memory  |

## Scope: identical reads only

|                                          |                                                                                                  |
| ---------------------------------------- | ------------------------------------------------------------------------------------------------ |
| ✅ Same key, same read (`GET /hot-item`) | collapse freely                                                                                  |
| ❌ Writes sharing a key                  | two distinct POSTs collapsed into one = a lost write                                             |
| ❌ Caller-cancelled leader               | leader's ctx dies → flight dies → followers fail together (scope ctx per flight, not per caller) |

## Cause → consequence

| ❌ Cause → consequence                                 | ✅ Instead                      |
| ------------------------------------------------------ | ------------------------------- |
| no singleflight → N misses, N queries, DB falls over   | one flight per key              |
| retry each follower → thunder again after shared error | retry wraps the flight once     |
| singleflight writes → lost updates                     | reads only, writes go direct    |
| per-request keys → zero sharing, memory grows          | key by resource, not by request |
