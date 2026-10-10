# Warehouse vs lakehouse — who holds the copy

Warehouse won the last decade on governance + easy SQL;
lakehouse rises on open formats giving warehouse features over
cheap storage: one copy, every engine.

|                 | Warehouse (Snowflake/BigQuery/Redshift) | Lakehouse (Delta/Iceberg + Trino/Spark)               |
| --------------- | --------------------------------------- | ----------------------------------------------------- |
| Data            | structured, schema-on-write             | all types, schema-on-read/evolution                   |
| Format          | proprietary, vendor-locked              | open (Parquet + Iceberg/Delta/Hudi) on object storage |
| Compute/storage | split, but inside the vendor            | fully split — any engine reads                        |
| Cost            | storage + query priced high             | cheap S3 storage, pay engines separately              |
| Governance      | built-in (RBAC, audit, sharing)         | assemble it (Iceberg REST catalog, policy engines)    |
| Fits            | standard BI, SQL team, govern fast      | big data + ML + BI on one copy, dodge lock-in         |
