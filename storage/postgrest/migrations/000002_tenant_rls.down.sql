DROP POLICY IF EXISTS tenant_isolation ON api.todos;
DROP POLICY IF EXISTS anon_full_access ON api.todos;
ALTER TABLE api.todos DISABLE ROW LEVEL SECURITY;
ALTER TABLE api.todos DROP COLUMN IF EXISTS tenant_id;
