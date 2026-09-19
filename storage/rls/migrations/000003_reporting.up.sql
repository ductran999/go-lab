-- System-admin analytics path: a dedicated read-only role that sees
-- across tenants. Explicit permissive policy (auditable) instead of
-- BYPASSRLS. Tenant roles are untouched.

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_roles WHERE rolname = 'report_reader'
    ) THEN
        CREATE ROLE report_reader NOLOGIN;
    END IF;
END
$$;

GRANT USAGE ON SCHEMA lab TO report_reader;
GRANT SELECT ON lab.documents TO report_reader;
GRANT SELECT ON lab.docs_big TO report_reader;

DROP POLICY IF EXISTS report_global_read ON lab.documents;
CREATE POLICY report_global_read ON lab.documents
    FOR SELECT TO report_reader
    USING (true);

DROP POLICY IF EXISTS report_global_read_big ON lab.docs_big;
CREATE POLICY report_global_read_big ON lab.docs_big
    FOR SELECT TO report_reader
    USING (true);
