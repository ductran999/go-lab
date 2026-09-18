-- Lab 03: owner/superuser bypass.
-- Run: psql -f queries/03-owner-bypass.sql
-- Expect: admin sees ALL rows regardless of app.tenant_id.
-- Lesson: RLS never applies to table owners and superusers.
-- Always test policies as the app role (SET ROLE lab_user).

-- No SET ROLE here: running as admin (owner + superuser).
SET app.tenant_id = 1;
SELECT id, tenant_id, body FROM lab.documents ORDER BY id;

RESET app.tenant_id;
