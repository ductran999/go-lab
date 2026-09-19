# RLS Trade-offs

**TL;DR:** RLS = strongest per-row guarantee, Postgres-only, policy bugs are silent. Pair with app scoping; measure before fearing overhead.

Keywords: portability, silent deny, debuggability, measured overhead.

## RLS vs alternatives

| # | Pattern | Guarantee | Portable | Failure mode |
| - | ------- | --------- | -------- | ------------ |
| 1 | App scoping (`WHERE` in code) | Discipline | Yes | One miss = leak, loud or silent |
| 2 | RLS (this lab) | Database | Postgres-only | Misconfig = silent deny (safe) or bypass (fatal) |
| 3 | Schema per tenant | Namespace | Postgres-only | Migration fan-out, pool complexity |
| 4 | DB per tenant | Infrastructure | Yes | Ops cost, connection sprawl |

## For RLS

- **Fail-closed default.** No policy match = no rows. Safer direction than app bugs (fail-open).
- **Framework-proof.** Any client (PostgREST, GORM, raw psql, BI tool) gets the same filter. No per-stack reimplementation.
- **Measured overhead ≈ 0.** Lab 05 (100k rows): RLS + composite index 0.08ms vs RLS-off 0.05ms. Policy inlines as `Index Cond`, no Filter node.
- **Audit-friendly.** Policies are schema: versioned in migrations, reviewable in one place.

## Against RLS

- **Postgres lock-in.** Policies don't travel to MySQL/another store. Migration cost is real.
- **Silent deny debuggability.** `UPDATE` touching 0 rows vs genuinely empty result look identical. Needs `EXPLAIN` literacy + policy-aware logging.
- **Superuser/owner blind spot.** Ops tooling running as owner sees everything; a compromised admin credential bypasses all of it.
- **Bulk/cross-tenant ops need escape hatches.** Analytics across tenants, tenant migration, backfills: either privileged role or `BYPASSRLS`, both must be gated and audited.
- **Connection pooler modes.** Transaction pooling + `SET LOCAL` is the safe combo (verified in GORM demo). Session pooling with `SET` leaks tenants.

## Verdict for this stack

- Default: RLS on tenant tables + app passes tenant explicitly (defense in depth, already implemented both sides).
- Skip RLS where: tables with no tenant concept (migrations, seeds), bulk analytics roles (dedicated bypass role, audited).
- Revisit at scale: partition by `tenant_id` when single-table size or skewed tenants break index plans.

## Beyond per-tenant: cron + analytics paths

- **Per-tenant cron** (health checks): loop tenants, one tx each with `SET LOCAL` (lab 07). RLS enforced, results attributable, no privilege escalation.
- **Global analytics** (system admin): dedicated read-only `report_reader` role with explicit permissive policy (lab 06). Operational and analytical paths stay separate; tenant roles never reach global aggregates.
- **Pre-aggregate at scale**: cron/ETL rollups into a reporting schema. Dashboards read rollups, never raw tenant tables on primary.

## Org cost: backend + DBA collaboration

- Policies live in the DB but encode product rules. Author needs both: backend (tenant model) + DBA (Postgres mechanics). Solo fullstack covers both; split teams need joint review.
- Review policies like code: versioned migrations, isolation matrix tests, never hand-edited on prod.
- Debugging crosses layers: app sees `[]`, cause lives in policy + plan. Teams without EXPLAIN literacy turn RLS into a black box.
- Deploys move in lockstep: new tenant dimension = migration + JWT claims + app code together. One side drifting breaks auth (or silently over-shares).
- Net: RLS trades technical bug surface for human coordination cost. Weak backend/DBA process favors app scoping + strict review instead.
