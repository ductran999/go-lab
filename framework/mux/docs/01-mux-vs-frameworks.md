# mux vs frameworks — control vs velocity

> TL;DR: labs use **mux** to see the guts; deadline CRUD uses
> **gin/echo** to ship. Pick by purpose, not religion.

## mux wins

- Zero deps, stable API, no magic — request path fully visible.
- Go 1.22+ closed the gap: method routing, `{wildcards}`,
  `{$}` anchors, `r.PathValue`.
- Middleware = function wrapping; no framework to learn.

## gin/echo win

- Binding + validation + error helpers in one line.
- Middleware ecosystem (CORS, JWT, rate-limit ready-made).
- Team velocity: familiar patterns, fast onboarding.
- Fiber (fasthttp) for raw RPS — when benchmarks demand it.

## Rule

- Learning/mechanics → mux. Shipping CRUD under deadline → framework.
