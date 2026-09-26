# OTel — three parts, one waterfall

> TL;DR: **SDK** creates spans, **propagation** (W3C headers)
> links them across hops, **collector** receives/batches/forwards,
> **Jaeger** draws. Sample in prod, keep all in labs.

## 1. Parts

- **SDK** (`internal/tracing`): tracer provider + OTLP exporter +
  W3C propagator, wired once per binary. Spans are cheap structs
  until exported; the batcher ships in the background.
- **Propagation**: `Inject` writes `traceparent` on the way out,
  `Extract` reads it on the way in. Same header as the trace lab —
  OTel just standardizes who writes it (otelhttp does this
  automatically; here manual to show the gears).
- **Collector**: vendor-neutral receiver (OTLP `:4317`) → batch →
  exporters (Jaeger, debug, Prometheus...). Backends change in YAML,
  code untouched.

## 2. Rules

- One trace per edge request; child spans per hop/operation
  (same split as trace/span/request_id in `../trace/`).
- Always-on sampling in labs; prod uses tail-based (keep errors +
  slow, drop the boring 99%).
- Async hops (queues) carry context in the message envelope —
  headers die at the broker (see `messaging/`).
