-- Lab 02: WITH CHECK on writes.
-- Run: psql -f queries/02-write-check.sql
-- Expect: own-tenant INSERT ok, cross-tenant INSERT 42501,
--          cross-tenant UPDATE touches 0 rows. No leftover rows.

SET ROLE lab_user;
SET app.tenant_id = 1;

-- Own tenant: allowed.
INSERT INTO lab.documents (tenant_id, body) VALUES (1, 't1-lab02')
RETURNING id, tenant_id, body;

-- Other tenant: denied (42501).
INSERT INTO lab.documents (tenant_id, body) VALUES (2, 'evil-lab02');

-- Cross-tenant UPDATE: matches 0 rows (silent deny).
UPDATE lab.documents SET body = 'hijacked' WHERE tenant_id = 2;

-- Cleanup own row. Find id first when running manually.
DELETE FROM lab.documents WHERE body = 't1-lab02';

RESET app.tenant_id;
RESET ROLE;
