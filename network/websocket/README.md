# WebSocket Origin Lab

WebSocket handshake carries `Origin`, but **no library checks it for
you**. `gorilla/websocket` upgrades anyone unless `CheckOrigin` says no.

```bash
# Change DIR
$ cd network/websocket
```

## Setup

```bash
make run-server   # API on :8094 (PORT to override)
make run-web      # demo page on :8093 (cross-origin vs :8094)
```

## Try it

```bash
# Evil origin -> 403 before upgrade (no socket opened)
curl -s -o /dev/null -w '%{http_code}\n' \
  -H 'Connection: Upgrade' -H 'Upgrade: websocket' \
  -H 'Sec-WebSocket-Version: 13' -H 'Sec-WebSocket-Key: x==' \
  -H 'Origin: http://evil.test' localhost:8094/ws
# 403

# Allowed origin -> 101 (curl hangs after: it can't speak frames)
curl -s -D- -o /dev/null --max-time 3 \
  -H 'Connection: Upgrade' -H 'Upgrade: websocket' \
  -H 'Sec-WebSocket-Version: 13' \
  -H 'Sec-WebSocket-Key: <base64-nonce>' \
  -H 'Origin: http://localhost:8093' localhost:8094/ws
# HTTP/1.1 101 Switching Protocols + Sec-WebSocket-Accept

# Browser on :8093 -> 101 Switching Protocols, echo works.
# Open http://localhost:8093, Connect, Send ping, watch echo.
```

## Dynamic origins + PNA

```bash
# Allowlist from env (default http://localhost:8093)
ORIGINS=http://localhost:8093,http://app.test make run-server
curl -s -o /dev/null -w '%{http_code}\n' -H 'Origin: http://app.test' \
  -H 'Connection: Upgrade' -H 'Upgrade: websocket' \
  -H 'Sec-WebSocket-Version: 13' -H 'Sec-WebSocket-Key: <base64-nonce>' \
  localhost:8094/ws
# 101 (curl hangs after: it can't speak frames)

# PNA preflight (simulated: public page → local socket)
curl -s -D- -o /dev/null -X OPTIONS localhost:8094/ws \
  -H 'Origin: https://public.app' \
  -H 'Access-Control-Request-Private-Network: true' | grep -iE 'HTTP|private-network'
# 204 + Access-Control-Allow-Private-Network: true
```

## Rules

- `CheckOrigin` allowlist, never `*` in prod. Default gorilla behavior
  (same-origin only) is safer than custom permissive funcs.
- Empty Origin (curl, non-browser) is rejected here: browsers always
  send it, so missing = suspicious. Adjust per use case.
- Origin check is necessary, not sufficient: pair with auth tokens
  on first message for anything sensitive.
