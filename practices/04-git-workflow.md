# Git workflow — small, topical, reversible

> TL;DR: **trunk-first**, short branches, **one topic per commit**,
> R&D dirty then commit clean. History must read like a changelog,
> not a diary.

## 1. Core idea

- `main` always releasable. Work on short branches (hours, not
  weeks), merge via PR with checks green (CI: vet/lint/tests).
- One commit = one topic (`feat: sse lab` ≠ `fix: lint` ≠ `docs:`).
  Reviewers read commits, not diffs.
- Commit **after** R&D settles (this repo's rule): spike messy,
  verify live, then stage topical commits. Never commit to
  "save progress" — that's what stashes are for.

## 2. Trunk vs gitflow (when each)

|          | Trunk-based              | Gitflow                                 |
| -------- | ------------------------ | --------------------------------------- |
| Branches | Short-lived, merge fast  | `develop`/`release`/`hotfix` long-lived |
| Releases | Continuous from main     | Versioned, scheduled                    |
| Fits     | SaaS, labs, small teams  | Shrink-wrap, multi-version support      |
| Cost     | Needs CI + feature flags | Merge trains, release overhead          |

- Default: trunk. Graduate to release branches only when two
  versions must live at once.

## 3. Rules applied here

- Separate `feat`/`fix`/`docs` commits (see log: range lab and
  secheaders fix went separately, never mixed).
- Amend only the tip for review fixups (`secheaders` note was
  amended pre-push); pushed history is append-only.
- `soft reset` to un-commit but keep work (SSE lab: R&D first,
  commit when nodded). Never rewrite shared history.
