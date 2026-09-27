# pgbouncer — 10 server conns serve 200 clients

Postgres `:5435` (`max_connections=10`) + pgbouncer `:6432`
(transaction mode, pool 10). Apps point at the pool, never the DB.
`make bench` (pgbench, 50 clients) shows direct choking vs pooled flat.

```bash
# Change DIR
$ cd storage/pgbouncer
```

## Run

```bash
make up            # db + pool
make init          # pgbench tables (once)
make appbench      # go clients direct vs pooled
make bench         # pgbench direct vs pooled (PGPASSWORD=password123)
make conns | make psql | make conf
make down
```

## Tests (no DB needed by default)

```bash
go test ./cmd/demo/ -v
# live direct instead:
PGDEMO_DSN='postgres://admin:password123@localhost:5435/pooldb?sslmode=disable' \
  go test ./cmd/demo/ -v
```

## Docs

- `docs/01-pooling.md` — why pool, 3 modes, what breaks
