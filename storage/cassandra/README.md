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

- `docs/00-cap-theorem.md` — CAP triangle, CP vs AP, the dial
- `docs/01-quorum.md` — RF/N/R/W, the W+R>RF rule, failure walk
- `docs/02-cassandra.md` — data modeling, consistency levels
- `docs/03-ring.md` — token ring, gossip, vnodes, graceful death
