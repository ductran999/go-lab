-- Lab 06: system-admin overall view across tenants.
-- Run: psql -f queries/06-global-report.sql
-- Expect: per-tenant counts plus grand total, all tenants visible.
-- Lesson: analytics travels a separate role/path from operational RLS.

SET ROLE report_reader;

-- Usage per tenant (the "how customers use the product" report).
SELECT tenant_id, count(*) AS docs FROM lab.documents GROUP BY tenant_id ORDER BY tenant_id;

-- Grand total.
SELECT count(*) AS total_docs FROM lab.documents;

-- Tenant admin contrast: same aggregate as lab_user sees only own tenant.
SET ROLE lab_user;
SET app.tenant_id = 1;
SELECT tenant_id, count(*) AS docs FROM lab.documents GROUP BY tenant_id ORDER BY tenant_id;

RESET app.tenant_id;
RESET ROLE;
