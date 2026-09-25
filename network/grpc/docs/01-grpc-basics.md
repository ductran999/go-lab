# gRPC basics — contracts that compile

> TL;DR: **protobuf** schema compiles to typed client+server,
> **HTTP/2** multiplexes calls on one connection, **4 call shapes**
> cover request/response and streams. Contract-first with teeth.

## 1. Wire

- **IDL**: `.proto` declares messages + service. `buf generate`
  emits Go structs, client, server interface — drift is a
  compile error, not a runtime surprise.
- **Encoding**: protobuf binary (small, fast, schema-bound).
  No human-readable curl; `grpcurl` + reflection fills the gap.
- **Transport**: HTTP/2 — one TCP connection, many concurrent
  streams (the 6-connection cap from the SSE lab retires),
  header compression, binary framing.

## 2. Four call shapes

| Shape            | Client → | Server →              | SSE/WS cousin  |
| ---------------- | -------- | --------------------- | -------------- |
| Unary            | 1 msg    | 1 msg                 | REST           |
| Server-streaming | 1 msg    | N msgs (`Watch` here) | SSE            |
| Client-streaming | N msgs   | 1 msg                 | Upload batches |
| Bidi             | N msgs   | N msgs                | WebSocket      |

## 3. Resume, typed

- `Watch{after_id}` replays missed events then tails live —
  same cursor contract as SSE `Last-Event-ID`, but the cursor
  is a typed field and the payload is a schema message.
- Slow watchers drop (buffered chan, same rule as every stream
  lab here): backpressure policy stays explicit, never silent OOM.

## 4. When (not) gRPC

- Service-to-service inside one org (typed, fast, codegen).
- Browser clients: grpc-web proxy needed (browsers don't speak
  raw gRPC) — public APIs usually stay REST/JSON.
- REST+H2 covers ~80%: public APIs, webhooks, CRUD — readable,
  debuggable, browser-native. gRPC pays off for internal meshes
  (compile-time contracts), big payloads (3–10x smaller), complex
  bidi streams, propagated deadlines. Rule: **outside → REST,
  inside → gRPC** (see `../http2/` for the REST-on-H2 proof).
- Ties to `practices/01-contract-first-vs-code-first.md`:
  gRPC is contract-first where the compiler enforces the contract.
