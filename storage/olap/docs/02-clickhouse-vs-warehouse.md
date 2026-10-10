# OLTP → OLAP → warehouse → lakehouse — pick by team, not tech

ClickHouse is not a basic RDBMS (that's row-based Postgres/MySQL
for OLTP) — it is a columnar OLAP engine, i.e. already a warehouse
engine. The real question is managed service vs self-run engine.

|         | OLTP (Postgres)        | OLAP engine (ClickHouse/DuckDB)  | Warehouse (Snowflake/BQ) | Lakehouse                        |
| ------- | ---------------------- | -------------------------------- | ------------------------ | -------------------------------- |
| Layout  | row                    | columnar                         | columnar, managed        | open Parquet                     |
| Good at | writes + point queries | giant scans + aggregations       | same, minus ops          | same + cheap + shared with ML    |
| Ops     | self/managed           | **self-run** (cluster, replicas) | vendor does all          | half (storage easy, engines DIY) |
| Cost    | low/mid                | low if you run it well           | high, pay per query      | cheapest storage                 |

| Pick              | When                                                                           |
| ----------------- | ------------------------------------------------------------------------------ |
| ClickHouse/DuckDB | huge scan volume + cost control + a team that runs clusters + realtime inserts |
| managed warehouse | no ops team, sharing/governance/elastic bursts, SQL-only team                  |
| lakehouse         | data lake exists + ML shares the copy                                          |
