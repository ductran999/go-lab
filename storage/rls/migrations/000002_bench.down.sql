DROP POLICY IF EXISTS tenant_isolation ON lab.docs_big;
ALTER TABLE lab.docs_big DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS lab.docs_big;
