# PK shapes — int vs uuid4 vs uuid7

> TL;DR: int fastest + smallest, u7 close behind, u4 slowest +
> fattest. At 50k the wall gap is modest (+14%) but the index gap
> is already ~2x — scatter compounds with size.

## Results (50k rows, local machine, 2nd run: warm cache)

|                          | Insert wall | pkey bytes | Deep page (`ORDER BY id OFFSET 40000 LIMIT 20`) |
| ------------------------ | ----------- | ---------- | ----------------------------------------------- |
| `t_int` (bigserial)      | 180ms       | 2,260,992  | 2.364ms                                         |
| `t_u4` (gen_random_uuid) | 182ms       | 4,374,528  | **9.685ms**                                     |
| `t_u7` (app v7)          | 182ms       | 3,178,496  | 2.603ms                                         |

## Why (read after running)

- int8 keys: 8 bytes, sequential → dense pages, minimal splits.
- uuid: 16 bytes always (2x key size before any scattering).
- u4 random: inserts land everywhere → page splits + WAL bloat.
- u7 ordered: appends at the edge → near-int locality.
- Read the numbers, not the theory: wall differs +14% but index
  differs 1.9x — buffer pool, backups and replicas all carry the
  fat index long after the insert benchmark ends.
- Inserts converge warm (all ~180ms); the durable gap is **reads**:
  deep pages on scattered u4 cost **4x** (9.7ms vs 2.4ms) — every
  page pays the sort tax forever. Ordered keys (int, u7) scan.
- Both uuids store **16 bytes** (the 36-char form is display only):
  u4's fatter index comes from **splits** (half-filled pages), not
  width. Same weight, different packing.
