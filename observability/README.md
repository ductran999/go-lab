# Observability — seeing the system

Traces, metrics, logs: one pillar for answering "what happened".
Seeded with distributed tracing; room reserved for Prometheus
(metrics), Loki (logs), Thanos (long-term).

## Map

| Dir | What |
|-----|------|
| `trace/` | W3C traceparent across svc-a → svc-b, trace vs span vs request_id |

Ties: `../network/grpc/` (interceptors log correlation IDs),
`../messaging/` (traceparent rides message envelopes, not headers).
