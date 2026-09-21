# Contract-first vs code-first

TL;DR: contract-first shines for public REST (spec is truth,
clients generate from it); code-first wins for speed and for
shapes specs cannot express (streams, internal APIs). Most teams
end up hybrid — split by API kind, not by dogma.

## 1. Core idea

- **Contract-first**: write the spec (`openapi.yaml`, GraphQL
  schema, protobuf) before code. Server and clients generate or
  conform to it. Review the contract in PRs like code.
- **Code-first**: write the server, derive the spec from
  annotations/reflection (`swaggo`, `drf-spectacular`). Spec
  follows implementation.

## 2. Trade-offs

|                         | Contract-first                                              | Code-first                              |
| ----------------------- | ----------------------------------------------------------- | --------------------------------------- |
| API design quality      | High (design before code)                                   | Drifts toward implementation-shaped     |
| Client codegen          | First-class                                                 | Works, spec may leak internals          |
| Review surface          | Small diff on yaml                                          | Full server diff                        |
| Speed to first endpoint | Slower                                                      | Faster                                  |
| Streams (SSE/WS)        | Not expressible (see `../network/docs/02-api-contracts.md`) | Same gap — spec is the limit either way |
| Internal APIs           | Needs spec-splitting discipline                             | Easy to accidentally publish            |

## 3. When to use which

- Public REST consumed by other teams/sdks → contract-first.
- Prototype, internal CRUD, single-team full-stack → code-first.
- Streams → neither: hand-write clients, share event structs.
- Hybrid (recommended): contract-first at the boundary that
  needs stability, code-first everywhere else.

## 4. Cousins (upcoming)

- `tdd.md` — test-first: same "write the contract early" instinct,
  applied to behavior instead of wire shape.
