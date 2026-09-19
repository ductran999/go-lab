-- Lab 07: per-tenant cron loop (health-check pattern).
-- Run: psql -f queries/07-tenant-loop.sql
-- Expect: each transaction sees exactly one tenant. RLS stays enforced;
-- no privileged role, no bypass. Each block = one cron tick per tenant.

-- Tick for tenant 1.
BEGIN;
SET LOCAL ROLE lab_user;
SET LOCAL app.tenant_id = 1;
SELECT count(*) AS tenant_1_docs FROM lab.documents;
COMMIT;

-- Tick for tenant 2.
BEGIN;
SET LOCAL ROLE lab_user;
SET LOCAL app.tenant_id = 2;
SELECT count(*) AS tenant_2_docs FROM lab.documents;
COMMIT;
