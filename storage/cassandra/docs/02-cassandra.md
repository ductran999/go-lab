# Cassandra — query first, tables second

> TL;DR: **partition key** decides _which node_, **clustering key**
> decides _order inside_. Design tables from queries, not entities.
> Consistency is a **dial** per query (ONE → QUORUM → ALL).
> Foundation: [`00-cap-theorem.md`](00-cap-theorem.md).

## 1. Data modeling (vs SQL instinct)

```text
keyspace lab
└── table timeline ── one partition per user ── rows sorted newest-first
    ┌──────────── partition user_id=7 ────────────┐
    │ created_at ↓ │ post_id │ body               │
    │ 10:00        │ aaa     │ "hello"            │
    │ 09:30        │ bbb     │ "world"            │  ← one read, in order
    └────────────────────────────────────────────┘
PRIMARY KEY ((user_id), created_at, post_id)
              ▲        ▲              ▲
              │        │              └── tiebreak, unique
              │        └── clustering: sort + range slice
              └── partition: which node holds it
```

| SQL instinct           | Cassandra rule                                         |
| ---------------------- | ------------------------------------------------------ |
| Normalize entities     | **Denormalize per query** (1 query = 1 table)          |
| JOIN at read           | Pre-join at write (duplicate data freely)              |
| WHERE anything (index) | WHERE must hit **partition key** (+ clustering prefix) |
| Sort in app            | `CLUSTERING ORDER BY` stores sorted                    |

- Our `timeline`: `PRIMARY KEY (user_id, created_at, post_id)` —
  all posts of a user live together, newest first, one partition read.
- Hot partition (celebrity user_id) = hotspot: split by time bucket
  (`(user_id, month)`) when one key outgrows a node.

## 2. Consistency dial (per query!)

```text
W + R > RF  → strong (quorum reads see quorum writes)
W=R=ONE     → fast, possibly stale (AP mode)
W=R=QUORUM  → balanced (typical prod: RF=3, QUORUM=2)
W=ALL/R=ALL → CP mode (a down node blocks writes AND reads)
```

- `CONSISTENCY QUORUM;` in cqlsh, or per-statement in drivers.
- Our lab RF=1: every level behaves like ONE (single copy!) —
  the dial matters at RF=3, noted here for when it does.

## 3. Rules

- No JOINs, no secondary-index fishing, no `ALLOW FILTERING`
  outside exploration (full-cluster scan in disguise).
- TTL columns for expiring data (`USING TTL 86400`) instead of
  DELETE jobs.
- Lightweight transactions (`IF NOT EXISTS`) are Paxos rounds:
  correct, ~4x latency — use sparingly.
