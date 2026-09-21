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

## Roadmap

- [x] `docs/01-request-journey.md` — full path with per-hop headers
- [x] `cors/` lab — hand-rolled middleware + preflight matrix (4 cases verified)
- [x] `websocket/` lab — handshake Origin check (403 evil/missing, browser 101)
- [x] `sse/` lab — streaming headers, kill+resume via Last-Event-ID, WS vs SSE doc
- [ ] Header deep-dives (cache, auth, SSE/streaming set)
- [ ] CORS lab (preflight matrix with curl)
- [ ] Trace propagation demo (2 services + shared trace ID end to end)
- [ ] `Last-Event-ID` resume + heartbeat tuning (feeds realtime lab)

Deep-dive notes in [`docs/`](docs/). Each doc: keywords, TL;DR, tables.
