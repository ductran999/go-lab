# UUID lab — v4 random, v7 time-ordered

Two UUIDs, two jobs. Run the demo, read the prefixes, pick by need —
then prove it with 50k inserts per PK shape.

```bash
# Change DIR
$ cd storage/uuid
```

## Run

```bash
make run  # v4 vs v7 side by side + order check (no DB needed)
```

## Bench (int vs uuid4 vs uuid7)

```bash
make up       # postgres :5438
make migrate  # t_int, t_u4, t_u7
make bench    # 50k rows each + index sizes
make down
```

## Docs

- `docs/01-uuid.md` — structure, trade-offs, when each
- `docs/02-pk-shapes.md` — bench results + why (filled after first run)
