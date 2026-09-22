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
- [x] `auth/` lab — one token three channels (header/cookie/query) + strict endpoint
- [x] `cache/` lab — ETag/304, max-age freshness, hits counter, http-caching doc
- [x] `trace/` lab — W3C traceparent across svc-a → svc-b, shared trace ID
- [x] `negotiation/` lab — Accept/q ranking, 406, Vary: Accept doc
- [x] `forwarding/` lab — XFF chain, trust by peer, spoof demo, blocklist verdict
- [x] `secheaders/` lab — bare vs hardened, polyglot + subresource nosniff demo
- [x] `range/` lab — 206/416 resume, parallel fetch client, checksum + hashproof
- [x] `load-balancing/` lab — six algorithms harness (moved from root)
- [ ] Header deep-dives (cache, auth, SSE/streaming set)
- [ ] CORS lab (preflight matrix with curl)
- [ ] Trace propagation demo (2 services + shared trace ID end to end)
- [ ] `Last-Event-ID` resume + heartbeat tuning (feeds realtime lab)

Deep-dive notes in [`docs/`](docs/). Each doc: keywords, TL;DR, tables.
