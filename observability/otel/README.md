# OTel end-to-end — grep graduation: waterfall in Jaeger

svc-a (`:8110`) starts a trace, calls svc-b (`:8111`); both export
OTLP to the collector, Jaeger draws the waterfall. `otelhttp`
handlers + transport do Extract/Start/Inject — our code reads spans,
never manages them. Ours to own: JWT auth, business attributes,
service identity.

```bash
# Change DIR
$ cd observability/otel
```

## Run

```bash
make up     # collector (:4317) + jaeger UI (:16686, in-memory)
# or: make up-with-badger  # persistent jaeger (UI :16688, OTLP :4319)
make run-b  # terminal 1
make run-a  # terminal 2
```

## Architecture: how traces reach Jaeger

```text
curl --(JWT)--> svc-a:8110 --(JWT+baggage)--> svc-b:8111
  | trace.id minted            | same trace.id (child span)
  | baggage{tenant,user} set   | baggage read, no re-auth
  | OTLP :4317                 | OTLP :4317
  v                            v
           collector:4317 (redact → batch)
                 |
          Jaeger :16686 (UI waterfall)
                 ^
  logs (slog JSON): trace.id/span.id/tenant.id/req.id ─┘ join key
```

- SDKs export OTLP/gRPC to the collector (never to Jaeger directly
  in the default path); the collector batches and forwards.
- Alternative: `jaeger-direct` profile (SDK → Jaeger `:4317`
  straight, UI `:16687`) — no collector, no buffer.
- One request = one log line per service, joined by `trace.id`
  (or `req.id` where traces don't reach).

```bash
# Mint once, reuse (1h expiry)
TOKEN=$(curl -s -X POST localhost:8110/token \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":"acme","user_id":"boss@company.com"}' | jq -r .token)

# Traced call (both services log auth + spans carry tenant/user)
curl -s localhost:8110/start -H "Authorization: Bearer $TOKEN" | jq

# open http://localhost:16686 → svc-a → one trace, two spans with
# tenant.id/user.id attributes
```

## Docs

- `docs/01-otel.md` — SDK, collector, propagation, sampling
