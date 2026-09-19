DROP POLICY IF EXISTS report_global_read_big ON lab.docs_big;
DROP POLICY IF EXISTS report_global_read ON lab.documents;
REVOKE ALL ON lab.docs_big FROM report_reader;
REVOKE ALL ON lab.documents FROM report_reader;
REVOKE USAGE ON SCHEMA lab FROM report_reader;
