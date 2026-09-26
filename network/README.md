# Network Lab

Research on what happens between client intent and server response:
request journey, headers, CORS, propagation IDs, tracing and
observability in distributed systems.

```bash
# Change DIR
$ cd network
```

## Scope

- Request journey: DNS → TCP/TLS → HTTP → load balancer → app → DB.
- Headers that matter: caching, auth, content negotiation, SSE/streaming.
- Cross-cutting IDs: `X-Request-ID`, trace/span propagation (W3C).
- CORS mechanics: preflight, credentials, common misconfigs.
- Observability: logs/metrics/traces joined by trace ID.

## Roadmap (grouped by layer)

- [x] `docs/` — request journey, API contracts
- [x] `http/` — cors, auth, cache, negotiation, forwarding, secheaders, range
- [x] `realtime/` — websocket (Origin/PNA), sse (headers/resume/usecases)
- [x] `rpc/` — grpc (unary/streams/interceptors), http2 (REST on H2), http3 (QUIC)
- [x] `infra/` — load-balancing harness + algorithms doc
- [x] ~~old flat items~~ — retired into groups above

> `trace/` moved to `../observability/trace/` (pillar seed).

Deep-dive notes in [`docs/`](docs/). Each doc: keywords, TL;DR, tables.
