# Ring — no master, just gossip and hashes

> TL;DR: every node is equal. **Murmur3** hashes partition keys
> onto a **token ring**; **gossip** spreads membership; **virtual
> nodes** balance the load. No master = no single point of death.

## 1. The ring

```text
token range: 0 ............ 2^63 ............ 2^64-1 (wraps)
             A ------ B ------ C ------ (back to A)

row user_id=7 → murmur3 → token T → first node clockwise from T owns it
(+ next RF-1 nodes clockwise hold copies)
```

- Writes/reads route by hash: any node coordinates (no master hop).
- Add a node: it takes token slices from neighbors, streams data over.

## 2. Gossip (membership without a boss)

- Every second, each node swaps state (up/down, load, schema version)
  with 1–3 random peers. Rumors converge in O(log N) rounds.
- **Phi-accrual failure detector**: tracks heartbeat arrival intervals
  statistically; phi > threshold = "probably dead" (adapts to slow
  networks instead of fixed timeouts).
- No consensus for membership — gossip is eventually consistent,
  good enough for routing (wrong guesses just retry elsewhere).

## 3. Virtual nodes (vnodes)

- One physical node holds **many small token ranges** (default 16–256
  vnodes) instead of one big arc: even spread + fast rebuild (a new
  node takes slivers from everyone, not half of one neighbor).
- Trade-off: more ranges = more metadata gossip; defaults are sane.

## 4. What dies gracefully

| Dies          | Happens                                              |
| ------------- | ---------------------------------------------------- |
| 1 node (RF=3) | QUORUM unaffected; hints queue for the dead          |
| Whole rack/DC | Survives if RF spans DCs (`NetworkTopologyStrategy`) |
| Coordinator   | Client retries; any node coordinates                 |
| Half the ring | QUORUM fails → writes refuse (CP moment inside AP)   |

## Glossary

| Word        | Means                                  | Example                         |
| ----------- | -------------------------------------- | ------------------------------- |
| gossip      | rumor-style peer state exchange        | Membership converges via gossip |
| vnode       | virtual token range on a physical node | 16 vnodes split load evenly     |
| coordinator | node handling your request             | Any node can coordinate         |
| phi-accrual | statistical death suspicion            | Phi > 8 ≈ node dead             |
