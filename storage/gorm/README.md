# gorm session lab — config per chain, DryRun visible

Same `*gorm.DB` pool underneath; `Session()` tweaks one chain.
All demos print SQL without executing (no live DB needed).

```bash
# Change DIR
$ cd storage/gorm
```

## Run

```bash
make up    # direct postgres :5436 (Open dials once, DryRun writes nothing)
make run
```

## Docs

- `docs/01-session.md` — flags, hooks order, pool relation
