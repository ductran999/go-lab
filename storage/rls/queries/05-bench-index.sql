-- Lab 05: index effect on RLS-filtered queries (100k rows, 10 tenants).
-- Run: psql -f queries/05-bench-index.sql
-- Stages: no index -> single index -> composite index -> RLS off compare.
-- Re-runnable (IF NOT EXISTS, ends with RLS enabled).
-- Observed (100k rows): stage 1-2 use pkey Index Scan + Filter
-- (LIMIT + ORDER BY id favors pkey order); stage 3-4 use composite
-- index with no Filter at all.

-- Stage 1: RLS on, no index. Expect Seq Scan.
SET ROLE lab_user;
SET app.tenant_id = 1;
EXPLAIN (ANALYZE, COSTS OFF)
SELECT id, body FROM lab.docs_big ORDER BY id LIMIT 100;
RESET ROLE;

-- Stage 2: single-column index. Expect Bitmap Heap Scan + Sort.
CREATE INDEX IF NOT EXISTS docs_big_tenant_id_idx ON lab.docs_big (tenant_id);
SET ROLE lab_user;
SET app.tenant_id = 1;
EXPLAIN (ANALYZE, COSTS OFF)
SELECT id, body FROM lab.docs_big ORDER BY id LIMIT 100;
RESET ROLE;

-- Stage 3: composite index. Expect Index Scan, no Sort node.
CREATE INDEX IF NOT EXISTS docs_big_tenant_id_id_idx ON lab.docs_big (tenant_id, id);
SET ROLE lab_user;
SET app.tenant_id = 1;
EXPLAIN (ANALYZE, COSTS OFF)
SELECT id, body FROM lab.docs_big ORDER BY id LIMIT 100;
RESET ROLE;

-- Stage 4: RLS off, explicit filter. Baseline for policy overhead.
ALTER TABLE lab.docs_big DISABLE ROW LEVEL SECURITY;
SET ROLE lab_user;
EXPLAIN (ANALYZE, COSTS OFF)
SELECT id, body FROM lab.docs_big WHERE tenant_id = 1 ORDER BY id LIMIT 100;
RESET ROLE;
ALTER TABLE lab.docs_big ENABLE ROW LEVEL SECURITY;
