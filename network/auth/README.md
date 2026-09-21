# Auth lab — one token, three channels

Page on `:8097`, API on `:8098`. Same HS256 token travels via
header, cookie, or query — `/me` reports which.

## Run

```bash
make run-server  # API on :8098
make run-web     # page on :8097
```

## Try

```bash
TOKEN=$(curl -s -XPOST localhost:8098/login \
  -H 'Content-Type: application/json' -d '{"user":"demo"}' | \
  python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')

# Header (the strict way)
curl -s localhost:8098/me -H "Authorization: Bearer $TOKEN"
# {"user":"demo","via":"header"}

# Query (works everywhere, leaks into logs)
curl -s "localhost:8098/me?token=$TOKEN"
# {"user":"demo","via":"query"}

# Strict endpoint rejects non-header channels
curl -s "localhost:8098/me-strict?token=$TOKEN"
# 401 missing token

# Tampered token
curl -s localhost:8098/me -H "Authorization: Bearer ${TOKEN}x"
# 401 invalid token signature
```

Browser on `:8097` → Login → each channel button shows its `via`.
Cookie button may be rejected under third-party-cookie blocking —
see docs.

## Docs

- `docs/01-auth-headers.md` — Bearer anatomy, channel trade-offs
