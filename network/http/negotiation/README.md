# Negotiation lab — one URL, many shapes

API on `:8102`. `/data` serves JSON, CSV, or plain text picked by
the `Accept` header. Absent header → JSON; unmatchable → 406.

## Run

```bash
make run-server  # API on :8102
```

## Try

```bash
# Default (no Accept) → JSON
curl -s localhost:8102/data

# CSV for spreadsheets/scripts
curl -s localhost:8102/data -H 'Accept: text/csv'

# q-values rank preferences: CSV wins here
curl -s localhost:8102/data \
  -H 'Accept: application/json;q=0.5, text/csv;q=0.9'

# Unsupported → 406 with the supported list
curl -s -o /dev/null -w '%{http_code}\n' localhost:8102/data \
  -H 'Accept: application/xml'

# Vary is load-bearing: caches key on Accept
curl -s -D- -o /dev/null localhost:8102/data | grep -i '^vary'
```

## Docs

- `docs/01-content-negotiation.md` — Accept/q/Vary mechanics
