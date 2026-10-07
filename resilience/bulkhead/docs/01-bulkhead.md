# Bulkhead — isolate failure

Essence: isolate failure — one sick downstream must not take the
healthy ones with it. Compartments (own workers, connections,
queues per downstream): B exhausts its own lane, C never notices.

## Mechanism

| Piece               | Does                                                               |
| ------------------- | ------------------------------------------------------------------ |
| pool per downstream | workers/connections split by dependency, never shared              |
| bounded queue       | overflow fails fast instead of queueing forever                    |
| isolation levels    | thread pool (coarse) → connection pool → semaphore per call (fine) |

## Tradeoff

|                           |                                                                 |
| ------------------------- | --------------------------------------------------------------- |
| ✅ Blast radius contained | B's outage spends B's budget only                               |
| ✅ Predictable capacity   | each downstream's max cost known upfront                        |
| 🪙 Partition waste        | idle B pool can't lend to busy C — utilization drops for safety |
| 🪙 Sizing is a bet        | too small throttles healthy traffic, too big is no bulkhead     |

## Cause → consequence

| ❌ Cause → consequence                       | ✅ Instead                        |
| -------------------------------------------- | --------------------------------- |
| shared pool → one slow dep parks all workers | pool per downstream               |
| unbounded queue → memory grows, latency lies | bounded queue, reject overflow    |
| no isolation → deploy of B brown-outs C      | bulkheads follow dependency lines |
