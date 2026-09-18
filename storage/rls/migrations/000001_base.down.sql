DROP POLICY IF EXISTS tenant_isolation ON lab.documents;
ALTER TABLE lab.documents DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS lab.documents;
DROP FUNCTION IF EXISTS lab.current_tenant();
DROP SCHEMA IF EXISTS lab;
