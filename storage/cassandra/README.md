# Cassandra lab — model by query, tune by consistency

Wide-column store: partition key spreads, clustering key orders.
No JOINs, no ad-hoc WHERE — table shape IS the query plan.

```bash
# Change DIR
$ cd storage/cassandra
```

## Run

```bash
make up     # :9042 (first boot ~1-2min: JVM + gossip; wait for healthy)
docker compose ps   # STATUS must show "healthy", not "starting"
make schema # apply schema.cql (needs healthy first!)
make run    # write QUORUM + read ONE demo
make cql    # open cqlsh for ad-hoc queries
```

## Docs

- `docs/01-cassandra.md` — data modeling, consistency levels
