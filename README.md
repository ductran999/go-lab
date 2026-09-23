# go-lab

![Go](https://img.shields.io/badge/go-1.26-blue) ![CI](https://github.com/ductran999/go-lab/actions/workflows/ci.yml/badge.svg)

Research playground. Each folder is one technology deep-dive with runnable
code, labs, and notes. Go is the primary implementation language, but any
tech that serves the research belongs here (SQL, Docker, frontend, ...).

| Folder | Research |
| ------ | -------- |
| `security/` | Identity + hashing: auth flows, mTLS, password/crypto hashes |
| `network/` | HTTP(S): CORS, WebSocket, SSE, auth, cache, trace, negotiation, forwarding, secheaders, range, load-balancing |
| `practices/` | Methodology: contract-first, TDD, clean architecture, git workflow |
| `storage/postgrest/` | PostgREST: REST from Postgres, JWT, multi-tenant RLS |
| `storage/rls/` | Postgres Row Level Security: policies, benchmarks, Go integration |
| `storage/realtime/` | Postgres NOTIFY → SSE bridge |

Conventions: English docs/comments, `golangci-lint` clean (`default: all`),
verify by running (`make tidy/vet/lint/build` at root).
