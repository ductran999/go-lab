# RLS Lab

Standalone Row Level Security experiments on Postgres 16.
No PostgREST, no JWT plumbing: tenant context travels in a session
setting (`SET app.tenant_id = N`), policies read it via
`lab.current_tenant()`.

Deep-dive notes in [`docs/`](docs/): how RLS works, the architecture
it solves, trade-offs with measured numbers.

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

| #   | File                             | Question                                 | Expected                                                        |
| --- | -------------------------------- | ---------------------------------------- | --------------------------------------------------------------- |
| 1   | `queries/01-basic-isolation.sql` | Read isolation per tenant?               | T1 sees 2 rows, T2 sees 1, unset sees 0                         |
| 2   | `queries/02-write-check.sql`     | `WITH CHECK` on writes?                  | Own INSERT ok, cross INSERT `42501`, cross UPDATE 0 rows        |
| 3   | `queries/03-owner-bypass.sql`    | Who bypasses RLS?                        | Owners + superusers see all rows always                         |
| 4 | `queries/04-explain.sql` | Policy cost in plan? | `Filter` with policy expression inline |
| 5 | `queries/05-bench-index.sql` | Index effect on RLS queries (100k rows)? | Composite `(tenant_id, id)` wins; RLS overhead ~nil (see below) |
| 6 | `queries/06-global-report.sql` | System-admin overall view? | `report_reader` aggregates all tenants; tenant role sees own only |
| 7 | `queries/07-tenant-loop.sql` | Per-tenant cron without bypass? | One tx per tenant via `SET LOCAL`, RLS stays enforced |

## Lab 05 results (100k rows, 10 tenants, `ORDER BY id LIMIT 100`)

| Stage                                  | Plan                                             | Time    |
| -------------------------------------- | ------------------------------------------------ | ------- |
| 1. RLS on, no tenant index             | pkey Index Scan + Filter (900/1000 rows removed) | ~0.95ms |
| 2. RLS on, single index                | Same (planner prefers pkey order for LIMIT)      | ~0.18ms |
| 3. RLS on, composite `(tenant_id, id)` | Index Scan, no Filter node                       | ~0.08ms |
| 4. RLS off, explicit `WHERE`           | Same as stage 3                                  | ~0.05ms |

- Lesson: policy inlines as a plain qual — with the right index it
  disappears from the plan entirely. RLS overhead ≈ noise.
- Surprise: `LIMIT` + `ORDER BY id` made the planner prefer pkey-order
  scan over bitmap, even with a single-column tenant index. Shape of
  query decides, not just indexes present.

Run inside `make psql` (`./queries` is mounted at `/labs/queries`):

```sql
\i /labs/queries/01-basic-isolation.sql
```

## Go demo: scoped vs RLS side by side

Two entry points, one per implementation. Each is self-contained:
read top to bottom to see the whole approach.

- `cmd/scoped`: explicit `Where("tenant_id = ?", ...)` in code.
- `cmd/rls`: bare `Find` in a tx with `SET LOCAL ROLE` +
  `SET LOCAL app.tenant_id`; the policy filters rows.

```bash
make run-scoped   # SELECT * ... WHERE tenant_id = 1 (2 rows)
make run-rls      # SELECT * (2 rows, no WHERE; RLS filtered)
make run          # both back-to-back
```

Same results, different enforcement point. `SET LOCAL`
keeps role and setting transaction-scoped, so pooled connections never
leak tenants.

## HTTP server: same comparison over REST

```bash
make run-server   # :8081
curl -s localhost:8081/scoped/tenants/1/docs | jq
TOKEN=$(make -s token TENANT=2)
curl -s localhost:8081/rls/docs -H "Authorization: Bearer $TOKEN" | jq
curl -s localhost:8081/rls/docs                        # 401, token required
```

- `/scoped/tenants/:id/docs`: tenant explicit in path (scoping demo).
- `/rls/docs`: tenant from JWT via middleware, never in params.
  Token: `make token TENANT=N` (HS256, `RLS_JWT_SECRET`).
- Responses are raw JSON arrays (`[]` when empty, never `null`).

## Cleanup

```bash
make down    # stop, keep data
make clean   # stop + remove postgres_rls_data volume
```

## Reading EXPLAIN (cheat notes)

```sh
                                                QUERY PLAN
----------------------------------------------------------------------------------------------------------
 Limit (actual time=0.040..0.095 rows=100 loops=1)
   ->  Index Scan using docs_big_tenant_id_id_idx on docs_big (actual time=0.038..0.081 rows=100 loops=1)
         Index Cond: (tenant_id = 1)
 Planning Time: 0.550 ms
 Execution Time: 0.137 ms
(5 rows)

```

- **Node** = operator in plan tree (Scan, Filter, Sort, Limit), not a server. Data flows bottom-up.
- **`actual time=A..B`** = ms to first row .. ms to all rows. Gap = rows scanned but discarded.
- **`loops=N`** = node executions. Times/rows are per-loop averages; multiply for total. High loops = nested-loop smell.
- **`Rows Removed by Filter: 900`** = read 1000, kept 100. Core waste metric.
- **Filter vs Index Cond**: Filter checks rows after read; Index Cond seeks matches directly. Policy inlined as Index Cond = RLS overhead ≈ 0.
- **COSTS OFF** hides planner guesses, keeps actuals. Costs are for the planner, not humans.
- **Planning Time** = parse + rewrite (RLS inline here) + optimize. Measured, one-time per query. RLS taxes planning, not execution.
- **End-to-end** = Planning + Execution + Transfer + Client decode. `EXPLAIN` covers server only; `curl -w time_total` covers all.
