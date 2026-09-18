-- Lab 01: basic read isolation.
-- Run: psql -f queries/01-basic-isolation.sql
-- Expect: tenant 1 sees 2 rows, tenant 2 sees 1 row, unset sees 0 rows.

SET ROLE lab_user;

SET app.tenant_id = 1;
SELECT id, tenant_id, body FROM lab.documents ORDER BY id;

SET app.tenant_id = 2;
SELECT id, tenant_id, body FROM lab.documents ORDER BY id;

RESET app.tenant_id;
SELECT id, tenant_id, body FROM lab.documents ORDER BY id;

RESET ROLE;
