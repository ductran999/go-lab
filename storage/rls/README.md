# RLS Lab

Standalone Row Level Security experiments on Postgres 16.
No PostgREST, no JWT plumbing: tenant context travels in a session
setting (`SET app.tenant_id = N`), policies read it via
`lab.current_tenant()`.

```bash
# Change DIR
$ cd storage/rls
```

## Setup

```bash
cp .env.example .env   # edit if needed
make up                # postgres:5433 + migrate up
make psql              # shell as admin (bypasses RLS!)
```

- Postgres: `localhost:5433` (host port differs from the postgrest lab).
- Dataset: `lab.documents`, 3 seed rows across tenants 1 and 2.
- Test role: `lab_user`. Admin is superuser and bypasses RLS —
  every lab starts with `SET ROLE lab_user`.

## Labs

| #   | File                             | Question                   | Expected                                                 |
| --- | -------------------------------- | -------------------------- | -------------------------------------------------------- |
| 1   | `queries/01-basic-isolation.sql` | Read isolation per tenant? | T1 sees 2 rows, T2 sees 1, unset sees 0                  |
| 2   | `queries/02-write-check.sql`     | `WITH CHECK` on writes?    | Own INSERT ok, cross INSERT `42501`, cross UPDATE 0 rows |
| 3   | `queries/03-owner-bypass.sql`    | Who bypasses RLS?          | Owners + superusers see all rows always                  |
| 4 | `queries/04-explain.sql` | Policy cost in plan? | `Filter` with policy expression inline |
| 5 | `queries/05-bench-index.sql` | Index effect on RLS queries (100k rows)? | Composite `(tenant_id, id)` wins; RLS overhead ~nil (see below) |

## Lab 05 results (100k rows, 10 tenants, `ORDER BY id LIMIT 100`)

| Stage | Plan | Time |
| ----- | ---- | ---- |
| 1. RLS on, no tenant index | pkey Index Scan + Filter (900/1000 rows removed) | ~0.95ms |
| 2. RLS on, single index | Same (planner prefers pkey order for LIMIT) | ~0.18ms |
| 3. RLS on, composite `(tenant_id, id)` | Index Scan, no Filter node | ~0.08ms |
| 4. RLS off, explicit `WHERE` | Same as stage 3 | ~0.05ms |

- Lesson: policy inlines as a plain qual — with the right index it
  disappears from the plan entirely. RLS overhead ≈ noise.
- Surprise: `LIMIT` + `ORDER BY id` made the planner prefer pkey-order
  scan over bitmap, even with a single-column tenant index. Shape of
  query decides, not just indexes present.

Run inside `make psql` (`./queries` is mounted at `/labs/queries`):

```sql
\i /labs/queries/01-basic-isolation.sql
```

## Cleanup

```bash
make down    # stop, keep data
make clean   # stop + remove postgres_rls_data volume
```

## Reading EXPLAIN (cheat notes)

- **Node** = operator in plan tree (Scan, Filter, Sort, Limit), not a server. Data flows bottom-up.
- **`actual time=A..B`** = ms to first row .. ms to all rows. Gap = rows scanned but discarded.
- **`loops=N`** = node executions. Times/rows are per-loop averages; multiply for total. High loops = nested-loop smell.
- **`Rows Removed by Filter: 900`** = read 1000, kept 100. Core waste metric.
- **Filter vs Index Cond**: Filter checks rows after read; Index Cond seeks matches directly. Policy inlined as Index Cond = RLS overhead ≈ 0.
- **COSTS OFF** hides planner guesses, keeps actuals. Costs are for the planner, not humans.
- **Planning Time** = parse + rewrite (RLS inline here) + optimize. Measured, one-time per query. RLS taxes planning, not execution.
- **End-to-end** = Planning + Execution + Transfer + Client decode. `EXPLAIN` covers server only; `curl -w time_total` covers all.
