# WAL shipping vs logical log — clone the machine vs compile the changes

| | WAL shipping (physical) | Logical log |
|---|---|---|
| Copies | disk block bytes, replayed identically | rows/changes (+ selected DDL) |
| Examples | Postgres streaming, Patroni standby | logical decoding, Debezium, binlog ROW |
| Versions | same major + arch, all-or-nothing cluster | heterogeneous (PG 14 → 16, PG → Kafka/warehouse) |
| Scope | whole cluster | per-table/row publications, filterable |
| Lag | byte-lag, fast replay | parse + apply per row — heavier under write storms |
| Fits | ✅ HA failover (sync standby promotes with zero loss) | ✅ zero-downtime migrate, CDC to warehouse, geo-sharding |
| Verdict | same engine, survive death | cross engines, move and reshape |
| 🪙 Footgun | lagging standby replays forever, never rots the primary | dead subscriber + live slot = WAL piles up → disk full → primary down with it |

## Consensus vs logical: two different questions

| Layer | Answers | Needs |
|---|---|---|
| consensus (Raft/Paxos) + WAL | dead → who takes over (failover, election) | identical appliers, byte for byte |
| logical log | data goes where next, becomes what (migrate, ETL, transform) | schema to read and reshape — never engine lock-in |

Production runs both: physical for in-AZ HA, logical for CDC
out to warehouse and Kafka.

## Two more flavors

| | Statement-based | Trigger-based |
|---|---|---|
| Copies | SQL text (MySQL STATEMENT binlog) | app triggers writing outbox/shadow tables |
| ✅ | tiny log, human-readable | no engine lock-in, ships with app code, selective by design |
| 🪙 | nondeterminism breaks replicas: `now()`, `uuid()`, `rand()`, bigserial order, `LIMIT` without `ORDER BY`, tz drift — MySQL moved to ROW for exactly this | trigger bugs are data bugs (silent skip, double-fire), write tax per row, schema changes break triggers, painful to observe |
| Verdict | legacy — readable but fragile | flexible — you own every bug (outbox pattern is its disciplined child) |
| 🪙 Footgun | replay diverges silently on nondeterminism — replicas disagree with no error | PII rides readable rows to warehouse/Kafka unless filtered (column lists, SMT masks); backfill new subscribers by hand |
