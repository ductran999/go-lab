# Trace lab — one ID across two services

svc-a on `:8100` (entry), svc-b on `:8101` (downstream). W3C
`traceparent` header carries the trace; both logs share one ID.

## Run

```bash
make run-b  # downstream :8101
make run-a  # entry :8100
```

## Try

```bash
# Fresh trace: a mints the ID, b inherits it
curl -s localhost:8100/start
# {"service":"a","trace_id":"<T>","downstream":{"service":"b","trace_id":"<T>"}}
# same <T> in both service logs

# Join an existing trace: pass traceparent in
curl -s localhost:8100/start \
  -H 'traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01'
# trace_id echoes 4bf92f3577b34da6a3ce929d0e0e4736 end to end

# Corrupt header: never fails, fresh trace starts
curl -s localhost:8100/start -H 'traceparent: garbage' | python3 -c \
  'import sys,json; print(json.load(sys.stdin)["trace_id"])'
```

## Docs

- `docs/01-trace-propagation.md` — traceparent anatomy, child spans
