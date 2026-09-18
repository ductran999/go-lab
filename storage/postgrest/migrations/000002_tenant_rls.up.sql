-- Multi-tenant hardening: required tenant_id + Row Level Security.
-- web_anon keeps an explicit open policy (lab convenience);
-- tenant_user is isolated per tenant via JWT claims (see 02-auth-model.md).

ALTER TABLE api.todos ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE api.todos ALTER COLUMN tenant_id DROP DEFAULT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_roles WHERE rolname = 'tenant_user'
    ) THEN
        CREATE ROLE tenant_user NOLOGIN;
    END IF;
END
$$;

GRANT tenant_user TO authenticator;

GRANT USAGE ON SCHEMA api TO tenant_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON api.todos TO tenant_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA api TO tenant_user;

ALTER TABLE api.todos ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS anon_full_access ON api.todos;
CREATE POLICY anon_full_access ON api.todos
    FOR ALL TO web_anon
    USING (true)
    WITH CHECK (true);

DROP POLICY IF EXISTS tenant_isolation ON api.todos;
CREATE POLICY tenant_isolation ON api.todos
    FOR ALL TO tenant_user
    USING (tenant_id = ((current_setting('request.jwt.claims', true)::json)->>'tenant_id')::int)
    WITH CHECK (tenant_id = ((current_setting('request.jwt.claims', true)::json)->>'tenant_id')::int);
