# Pooling — Postgres forks per connection, poolers don't

> TL;DR: Postgres **forks a process per connection** (~10MB each,
> max_connections wall). pgbouncer multiplexes **many clients →
> few server conns**. Point apps at `:6432`, never `:5435`.

## 1. Why

- 200 app conns direct = 200 backends ≈ 2GB RAM + fork storms.
  Same 200 through the pool = 20 server conns, rest wait fairly.
- Connection setup (fork + auth + TLS) costs ms each: pooled
  conns are warm, short queries skip the handshake tax.

## 2. Three modes (pick by workload)

| Mode          | Server conn held     | SET/transaction state | Use when                          |
| ------------- | -------------------- | --------------------- | --------------------------------- |
| `session`     | Whole client session | Safe (sticky)         | Long sessions, legacy apps        |
| `transaction` | Per transaction      | **Reset each tx**     | Stateless APIs (default pick)     |
| `statement`   | Per statement        | Nothing survives      | Autocommit single statements only |

## 3. What breaks (transaction mode gotchas)

- `SET LOCAL` / `SET` leaks? No — reset per tx, but **session-level
  SET vanishes**: our RLS lab's `SET LOCAL` inside explicit
  transactions still works; `SET` outside tx does not.
- Prepared statements: server-side `PREPARE` dies with the checkout
  → use `max_prepared_statements` or client-side prepares (pgx does).
- LISTEN/NOTIFY: needs a sticky conn — our realtime listener must
  bypass the pool (dedicated direct conn, as documented).
- Advisory locks: released at tx end — don't span transactions.

## 4. Rules

- Pool size ≈ `(2 × CPU) + disks` for server conns; clients 10x that.
- One pool per role/db; `pool_mode` per workload, not per server.
- Bench it (`make bench`): direct collapses past ~100 conns,
  pooled stays flat — numbers beat theory.

## 5. PoC: app pool ≠ server pool (20 clients, 10 slots)

| | Direct `:5435` | Pooled `:6432` |
|---|---|---|
| Server conns | 10 (`max_connections`) | **5** (`DEFAULT_POOL_SIZE`) |
| App clients | 20 | 20 |
| Result | Fails `53300 too many clients` | Passes clean |
| Proved by | `TestTooManyClients` | `TestPooledSurvives` |

- pgxpool multiplexes goroutines over *its own* conns, but each
  pool conn still costs **one server backend**: 20 app conns need
  20 server slots, pool or not. App-side pooling saves handshake
  churn, never server slots.
- Only a server-side pooler (transaction mode) breaks the 1:1:
  5 backends serve 20 clients. Two pools, two jobs — stack them.
- HTTP `429` is the API layer's mapping of `53300`, not PG's doing.
