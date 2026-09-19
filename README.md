# go-lab

Research playground. Each folder is one technology deep-dive with runnable
code, labs, and notes. Go is the primary implementation language, but any
tech that serves the research belongs here (SQL, Docker, frontend, ...).

| Folder | Research |
| ------ | -------- |
| `load-balancing-alg/` | Load balancing algorithms with live backends |
| `sse/` | Server-Sent Events end to end |
| `storage/postgrest/` | PostgREST: REST from Postgres, JWT, multi-tenant RLS |
| `storage/rls/` | Postgres Row Level Security: policies, benchmarks, Go integration |

Conventions: English docs/comments, `golangci-lint` clean (`default: all`),
verify by running (`make tidy/vet/lint/build` at root).
