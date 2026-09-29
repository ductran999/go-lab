# Quorum — majority math that makes consistency tunable

> TL;DR: with RF=3, **2 acks** make a write durable-enough and a
> read fresh-enough. The rule: **W + R > RF** guarantees overlap.
> Everything else is details.

## 1. Vocabulary (RF, N, R, W)

- **RF (replication factor)**: copies of each row. Lab: RF=1
  (single node — every level behaves like ONE). Prod: RF=3.
- **W**: replicas that must ack a write before success.
- **R**: replicas consulted on a read (newest timestamp wins).
- **ONE / QUORUM / ALL**: W or R = 1 / majority / all replicas.
  Majority of 3 = 2 (`floor(3/2) + 1`).

## 2. The write path (W=QUORUM, RF=3)

```text
client → coordinator → all 3 replicas
           ├── replica A: ack ✓
           ├── replica B: ack ✓  → 2 acks = QUORUM → success to client
           └── replica C: slow/down → hint stored, delivered later
```

- Client waits for **2 round trips** (paced by the slower one):
  slower than ONE, faster than ALL, survives 1 node down.
- Missed replica heals via **hinted handoff** (coordinator replays)
  or **read repair** (a later read spots the stale copy, fixes it).

## 3. The inequality (why it works)

```text
W + R > RF  →  read-quorum ∩ write-quorum ≠ ∅
2 + 2 > 3   →  at least 1 node has the latest data. Always.
```

- QUORUM+QUORUM: strong (within a healthy cluster).
- ONE+ONE: fast, maybe stale (different single nodes).
- ALL+ALL: strictest — and any 1 node down blocks everything.

## 4. Failure walk (RF=3, split 2-vs-1)

|                   | W=ONE                | W=QUORUM   | W=ALL      |
| ----------------- | -------------------- | ---------- | ---------- |
| Majority side (2) | succeeds             | succeeds   | **fails**  |
| Minority side (1) | succeeds (diverges!) | **fails**  | **fails**  |
| After heal        | needs repair         | consistent | consistent |

- QUORUM refuses the minority: no divergence, no repair bill.
- ONE accepts everywhere: always up, repair bill on heal.

## Glossary

| Word                | Means                              | Example                          |
| ------------------- | ---------------------------------- | -------------------------------- |
| quorum (majority)   | more than half                     | QUORUM of 3 is 2                 |
| hinted handoff      | coordinator replays missed writes  | Node returns, hints delivered    |
| read repair         | a read fixes stale copies it meets | Client triggers healing for free |
| overlap (intersect) | shared node between quorums        | W+R>RF forces an overlap         |
