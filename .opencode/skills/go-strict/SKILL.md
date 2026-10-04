---
name: go-strict
description: Use when writing or changing Go code, or before commit/push. Strict lint rules plus the local gate (vet, lint, test, gitleaks).
---

# Go strict

`default: all` lint — write it right the first time.

1. Errors: sentinel vars + `%w` wrap (`err113`). Never inline:
   declare `err` first, then a cuddled `if err != nil`
   (`noinlineerr`, `wsl`). Check every return, including
   `Close` in defers (`_ =` is a decision, not an accident).
2. Constructors return concrete types (`ireturn`); callers bind
   to narrow interfaces. Exported symbols get doc comments.
3. Structs public, fields private, built via `New` — misuse
   fails loud at construction, not silent downstream.
4. Docs and comments 100% English. No emojis unless asked.
5. Gate before push, in order: `gofmt -l .` (empty) → `make vet`
   → `make lint` (`lint-fix` only for its own diffs, re-read them)
   → `make test` → `gitleaks protect --staged`. Full sweep is
   `make check` (tidy, vet, build, test).
6. Never commit unless explicitly asked. Before any commit:
   `git status`, `git diff`, `git log --oneline -10`; stage only
   intended files; keep `go 1.26.0` and unrelated tidy churn out.
