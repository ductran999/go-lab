# OTel end-to-end — grep graduation: waterfall in Jaeger

svc-a (`:8110`) starts a trace, calls svc-b (`:8111`) with W3C
injection; both export OTLP to the collector, Jaeger draws the
waterfall. Manual Extract/Inject (no otelhttp) to show the mechanism.

```bash
# Change DIR
$ cd observability/otel
```

## Run

```bash
make up     # collector (:4317) + jaeger UI (:16686)
make run-b  # terminal 1
make run-a  # terminal 2
curl -s localhost:8110/start | jq
# open http://localhost:16686 → svc-a → one trace, two spans
```

## Docs

- `docs/01-otel.md` — SDK, collector, propagation, sampling
