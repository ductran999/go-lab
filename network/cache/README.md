# Cache lab — freshness vs validation

API on `:8099`. A `/hits` counter proves which requests reached
the origin at all. All demos are curl (caching is a header game).

## Run

```bash
make run-server  # API on :8099
```

## Try

```bash
# 1. Validation: revalidation reaches origin, saves bytes
curl -s -D- -o /dev/null 'localhost:8099/version?v=1' | grep -iE 'etag|cache-control'
# ETag: "1", Cache-Control: no-cache
curl -s -o /dev/null -w '%{http_code} %{size_download}\n' 'localhost:8099/version?v=1'
# 200 ~1040 bytes
curl -s -o /dev/null -w '%{http_code} %{size_download}\n' \
  -H 'If-None-Match: "1"' 'localhost:8099/version?v=1'
# 304 0 bytes (origin hit, body spared)
curl -s localhost:8099/hits  # counter climbed: 304s still cost a trip

# 2. Freshness: max-age avoids origin entirely (browser/proxy serve alone)
curl -s -D- -o /dev/null localhost:8099/fresh | grep -i cache-control
# public, max-age=60

# 3. Opt-out: nothing stored, every request is a trip
curl -s -D- -o /dev/null localhost:8099/nostore | grep -i cache-control
# no-store
```

Freshness in a real browser: open DevTools Network, hit `/fresh`
twice — second shows `(disk cache)`, `/hits` flat.

## Docs

- `docs/01-http-caching.md` — freshness, validation, SWR, Vary ties
