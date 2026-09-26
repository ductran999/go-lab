# Production concerns — interceptors, deadlines, health

> TL;DR: **interceptors** = middleware (log once, auth once),
> **deadlines** propagate and cancel work, **health** is a standard
> endpoint LB/K8s poll. Three lines of setup each, production-grade
> behavior.

## 1. Interceptors

- Unary chain: logging (method/duration/code) + auth (Bearer
  metadata, health/reflection skipped). Cross-cutting logic lives
  once, handlers stay pure business.
- Stream interceptor logs open/close only — per-message logging
  would drown the logs on chatty streams.

## 2. Deadlines + cancel

- Client: `Sleep{2000ms}` with 500ms deadline → `DeadlineExceeded`.
- Server: `select { <-ctx.Done() | <-timer.C }` — stops waiting
  instead of burning 1.5s of useless work. Cancel propagation is
  automatic over the wire; honoring it is the handler's job.
- Rule: every RPC gets a deadline (client sets, server respects).
  No deadline = one slow downstream hangs the chain forever.

## 3. Health

- Standard `grpc.health.v1`: `healthz.SetServingStatus(SERVING)`,
  LB/K8s poll it, rolling updates drain on NOT_SERVING.
- No token required (probes can't authenticate) — health reveals
  status only, never data.
