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

## Roadmap (grouped by API style)

- [x] `docs/` — request journey, API contracts
- [x] `headers/` — cors, auth, cache, negotiation, forwarding, secheaders, range
- [x] `rest/` — http2/http3 versions (H1 implicit everywhere)
- [x] `realtime/` — websocket (Origin/PNA), sse (headers/resume/usecases)
- [x] `rpc/` — grpc only (unary/streams/interceptors)
- [x] `infra/` — load-balancing harness + algorithms doc
- [x] ~~old flat items~~ — retired into groups above

> `trace/` moved to `../observability/trace/` (pillar seed).

Deep-dive notes in [`docs/`](docs/). Each doc: keywords, TL;DR, tables.
