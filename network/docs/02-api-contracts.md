# API contracts — where OpenAPI stops

TL;DR: OpenAPI + codegen own REST public APIs. Streams (SSE/WS)
and internal-only endpoints fall outside its shape — give each
its own tool instead of forcing one spec to cover everything.

## 1. OpenAPI's home turf

- Request/response REST, public audience, `openapi-codegen`
  generating typed clients from the spec (contract-first).
- One spec, one truth, compile-time breakage on drift.

## 2. Gap 1: streams have no shape in OpenAPI

- SSE documents as `GET → text/event-stream` with a string schema
  at best. Codegen awaits the full body — kills streaming,
  ignores `event:`/`id:`/`retry:`, no resume.
- WebSocket is not describable at all.
- Practice: keep REST in OpenAPI + codegen; hand-write stream
  clients (EventSource ≈ 10 lines). Share only the event JSON
  structs across the boundary.
- Already on GraphQL? Subscriptions are the typed answer:
  schema-declared streams over WS/SSE, `graphql-codegen` emits
  typed hooks, reconnect handled by client libs. Heavy unless
  GraphQL is already in the stack.

## 3. Gap 2: internal APIs need a different audience

- One published spec serves everyone; admin/debug/inter-service
  endpoints should not leak into the public document.
- Practice: split specs by audience (`public.yaml` vs
  `internal.yaml`, separate codegen/gateway), or keep internal
  out of OpenAPI entirely (gRPC + internal registry).

## 4. Rule of thumb

- REST public → OpenAPI + codegen.
- One-way push → SSE, hand-written client (`../sse/`).
- Chatty both-ways → WebSocket (`../websocket/`).
- Typed streams inside GraphQL → Subscriptions.
- Internal-only → separate spec or separate stack.
