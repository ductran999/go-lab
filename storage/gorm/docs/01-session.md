# gorm Session — one chain, one config

> TL;DR: `Session()` returns a **config-scoped handle** sharing the
> pool (0 connections). Flags decide hooks, transactions, guards —
> **before** any pool checkout happens.

## 1. Flags that matter

| Flag | Effect | When |
|------|--------|------|
| (default) | Writes wrapped in tx; hooks fire | Normal path |
| `SkipDefaultTransaction` | No explicit wrapper (implicit single-stmt tx remains) | Single-row writes, bulk import |
| `SkipHooks` | Before/After callbacks never run | Bulk ops, seeds |
| `AllowGlobalUpdate` | Permits WHERE-less Update/Delete | Migrations only (default block saves careers) |
| `DryRun` | Build SQL, skip execution | Previews, tests, this lab |
| `PrepareStmt` | Cache server prepares per conn | Hot repeated queries |
| `NewDB` | Fresh chain, drop prior clauses | Reusing a scoped handle safely |

## 2. Order of one query

```text
Session chain (hooks/callbacks, in-process)
  → pool checkout → pgbouncer assign → backend execute
```

Hooks run **before** the pool is touched: `DryRun` prints SQL with
no DB at all. Session = the director, pool = the stage.

## 3. Rules

- One behavior per chain: don't stack 5 flags, name the need.
- `AllowGlobalUpdate: true` never leaves a migration file.
- Prefer explicit tx (`db.Transaction`) over relying on the
  implicit default wrapping for multi-statement units.

## 4. Hooks ≈ triggers at app layer

| | PG trigger | GORM hooks |
|---|---|---|
| Runs in | DB (plpgsql) | App (Go) |
| Can do | Validate, audit, cascade data | + API calls, queues, metrics — anything Go does |
| Applies to | Every writer | Only this GORM app |
| Skippable | No (minus superuser) | `SkipHooks` flag |

- Integrity belongs in DB (no app bypasses it); side-effects
  belong in app (flexible). Need both, never substitute.
