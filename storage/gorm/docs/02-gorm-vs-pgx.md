# gorm vs pgx — ergonomics vs control

> TL;DR: **gorm** for velocity (models, hooks, migrations),
> **pgx** for control (pools, protocols, performance). gorm sits
> _on_ database/sql; pgx _is_ the driver. Know both, default by need.

|                 | gorm                                     | pgx (native)                                   |
| --------------- | ---------------------------------------- | ---------------------------------------------- |
| API             | Models + chains (`Find`, `Where`, hooks) | Raw SQL + args (`QueryRow`, `Exec`)            |
| Learning        | Fast start, magic underneath             | Explicit, nothing hidden                       |
| Hooks/callbacks | Built-in (BeforeCreate...)               | Hand-rolled                                    |
| Migrations      | AutoMigrate (dev)                        | golang-migrate / SQL files                     |
| Pool            | Via `database/sql` (generic)             | `pgxpool` (PG-aware: prepared cache, listen)   |
| Protocols       | Whatever database/sql negotiates         | Extended/simple per call, LISTEN/NOTIFY native |
| Performance     | Overhead per row (reflection)            | Fastest Go path to PG                          |
| Best for        | CRUD velocity, prototypes                | Hot paths, custom SQL, pool tuning             |

## Rules

- Prototype/CRUD → gorm (this lab). Hot loop/custom → pgx
  (pgbouncer demos use it for a reason).
- gorm `Session()` ≠ connection (config chain); pgx pool/wiring
  is explicit — re-read `../pgbouncer/docs/03-concepts.md` before
  arguing performance.
- Mixed is normal: gorm for 90% CRUD, raw pgx for the 10% hot.
