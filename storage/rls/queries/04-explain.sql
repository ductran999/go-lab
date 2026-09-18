-- Lab 04: policy overhead in the query plan.
-- Run: psql -f queries/04-explain.sql
-- Expect: Seq Scan with Filter showing the policy expression inline.
-- R&D: compare timing with RLS disabled vs enabled on larger datasets.

SET ROLE lab_user;
SET app.tenant_id = 1;

EXPLAIN (COSTS OFF) SELECT id, body FROM lab.documents ORDER BY id;

RESET app.tenant_id;
RESET ROLE;
