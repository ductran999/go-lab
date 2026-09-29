# UUID v4 vs v7 — random vs clock

> TL;DR: **v4** = 122 random bits (opaque, unguessable order).
> **v7** = Unix-ms + 74 random bits (sortable, leaks time).
> PKs/index keys → **v7**; unpredictability → **v4** (or secrets).

## 1. Structure (128 bits, `xxxxxxxx-xxxx-Mxxx-Nxxx-...`)

```text
v4: rrrrrrrr-rrrr-4rrr-yrrr-rrrrrrrrrrrr   (4 = version, y = variant)
v7: tttttttt-tttt-7ttt-yttt-tttttttttttt   (t = unix ms, 7 = version)
```

- v4: 122 random bits. Two v4s compare randomly — B-tree inserts
  scatter (page splits, fragmentation, bigger working set).
- v7: 48-bit ms timestamp + 74 random bits. Generated order ≈
  sorted order — append-mostly indexes (B-tree, LSM) stay compact.

## 2. Trade-offs

|                | v4                       | v7                                          |
| -------------- | ------------------------ | ------------------------------------------- |
| Order          | Random                   | Time-ordered                                |
| Index locality | Scatters (fragmentation) | Appends (compact)                           |
| Time leak      | None                     | Creation ms visible in prefix               |
| Uniqueness     | 122 random bits          | 74 random bits/ms (fine, plus monotonicity) |
| Collision      | Astronomically unlikely  | Same (per-ms randomness suffices)           |

## 3. Rules

- DB primary keys, event ids, anything indexed → v7.
- External tokens, anything where creation time must not leak →
  v4 (or opaque crypto-random strings, not UUIDs at all).
- Never v1 (MAC address leaks!) — v7 replaces it everywhere.
- Our `timeline.post_id`: v7 fits (time-ordered clustering key
  companion); v4 scatters for no benefit.

## Glossary

| Word           | Means                     | Example                   |
| -------------- | ------------------------- | ------------------------- |
| locality       | nearby keys stored nearby | v7 appends keep locality  |
| monotonic(ish) | non-decreasing order      | v7 sorts by creation time |
