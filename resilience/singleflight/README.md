# Singleflight lab — one key, one flight

Server on `:8123`. `/hot?ms=N` burns N ms per query and counts
executions, `/stats` shows the count, `/reset` zeroes it.

```
20 callers, same key "hot-item", -fly=true
  ├─▶ caller 1 (leader) ──▶ 1 origin query
  ├─▶ caller 2..20 (followers) ──▶ wait + share leader's answer
  └─▶ server sees: {"queries":1}   (without: {"queries":20})
```

## Run

```bash
make run-server  # expensive origin on :8123
make run-client  # 20 callers, collapsed (ARGS="-fly=false" to compare)
```

## Try

```bash
# 1. Stampede: 20 identical callers, 20 origin queries
curl -s localhost:8123/reset
go run ./cmd/client -n 20 -fly=false
# ok=20 shared=0 ... server={"queries":20}

# 2. Collapsed: 20 callers, 1 origin query, all share
curl -s localhost:8123/reset
go run ./cmd/client -n 20 -fly=true
# ok=20 shared=20 ... server={"queries":1}
```

## Docs

- `docs/01-singleflight.md` — mechanism, tradeoff, scope, cause
