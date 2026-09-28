# pgbadger lab — logs in, slow queries out

Own Postgres `:5437` with analysis logging baked in (slow 200ms+,
connects, attribution prefix). The loop: **traffic → logs →
report → tune**. Self-contained: revisit anytime, setup lives here.

```bash
# Change DIR
$ cd storage/pgbadger
```

## Run

```bash
make up      # postgres :5437
make load    # pgbench traffic (slow queries land in logs)
make fetch   # container logs → ./tmp/pg.log
make report  # → ./tmp/report.html, open in browser
make down    # stop (fetch first: logs die with the container!)
make clean
```

## Docs

- `docs/01-pgbadger.md` — what to log, report reading, tuning loop
