# Trace propagation — one ID, every hop

> TL;DR: **trace ID** is shared end to end, **span ID** is new per
> hop. `traceparent` carries both in **one header**; corrupt or
> absent → **start fresh**, never fail the request.

```mermaid
sequenceDiagram
    C->>A: GET /start (+ traceparent?)
    A->>B: GET /work + traceparent (same trace, new span)
    B-->>A: {trace_id: T}
    A-->>C: {trace_id: T, downstream: {trace_id: T}}
    Note over A,B: both logs grep-able by T
```

## 1. Header anatomy (W3C trace-context)

```text
traceparent: 00-<trace-id>-<parent-id>-<flags>
             │   32 hex      16 hex      01 = sampled
             └── version 00
```

- **trace-id**: minted once at the edge, immutable downstream.
- **parent-id**: this hop's span; each service mints a child
  span for the next call (`Child()` keeps trace, renews span).
- Malformed/absent → new trace. Propagation is best-effort
  decoration, never a gate.

## 2. Rules

- Read on entry, write on exit (every outbound call forwards).
- Log `trace_id` on every line that matters — correlation is
  `grep <T>` across services, no collector needed for the demo.
- Async work (queues, cron) carries the context in the message
  envelope — headers only survive HTTP hops.
- Production: forward to a collector (OTel SDK, Jaeger/Tempo);
  the header contract stays identical, only the sink changes.

## 3. trace vs span vs request_id

- **trace_id**: one end-to-end journey. Minted once at the edge,
  immutable to the end.
- **span_id**: one timed unit of work (a hop, a DB call). New
  span per network hop + per measurable operation; parent links
  form the waterfall tree under the trace.
- **request_id**: one HTTP request, minted at ingress/gateway,
  forwarded as-is. Logs-only correlation, no tree, no timing.
- Chaining `request_id` across hops is poor-man's tracing:
  enough for 2–3 linear services, blind at fan-out (parallel
  branches share one ID) and useless for latency attribution.
  Rule: linear + find-the-log → request_id; **which hop is
  slow** or fan-out → trace/span.
