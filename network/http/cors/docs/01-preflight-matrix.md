# Preflight Matrix

**TL;DR:** No Origin → no CORS. Simple + Origin → ACAO. Non-simple → OPTIONS first. No middleware → 404, real request never sent.

Keywords: preflight, ACAO, gates.

Curl flags used below:

- `-s` — silent: no progress meter, clean scriptable output.
- `-D-` — dump response **headers** to stdout (`-` = stdout): status line + headers.
- `-o /dev/null` — discard the response **body**.
- Combo `-D- -o /dev/null` = status + headers only, body skipped. Ideal for CORS checks.
- `-X OPTIONS` — force the HTTP method (default GET).
- `-H '...'` — attach a request header (`Origin`, preflight headers...).

Base: `http://localhost:8090`.

## 1. No Origin → no CORS involved

```bash
curl -s -D- localhost:8090/api/ping
# 200, no Access-Control-* headers. Same-origin / curl without
# Origin skips CORS entirely.
```

Annotated response:

```bash
HTTP/1.1 200 OK                               # HTTP version 1.1, status 200, reason phrase OK.
Content-Type: application/json; charset=utf-8 # Body is JSON encoded in utf-8.
Date: Sun, 20 Sep 2026 03:10:28 GMT           # Server timestamp when the response was created (not client-received time).
Content-Length: 18                            # Body is exactly 18 bytes. Mismatch means truncation (proxy cut, timeout mid-body).

{"message":"pong"}                            # Body, 18 bytes: matches Content-Length above.
```

## 2. Simple GET with Origin → ACAO echo (via nginx cache)

```bash
curl -s -D- -o /dev/null localhost:8085/api/ping -H 'Origin: http://app.test'
# 200 + Access-Control-Allow-Origin: * (needs make cache-up running)
```

Annotated response (full proxy path):

```bash
HTTP/1.1 200 OK                               # HTTP version 1.1, status 200, reason phrase OK.
Server: nginx/1.31.6                          # answered by proxy, not Go (server_tokens off in prod).
Access-Control-Allow-Origin: *                # any origin allowed (no credentials mode).
Content-Type: application/json; charset=utf-8 # Body is JSON encoded in utf-8.
Vary: Origin                                  # caches: key entries by Origin, don't share.
Date: Sun, 20 Sep 2026 04:01:16 GMT           # server origin time.
Content-Length: 18                            # body bytes, mismatch means truncation.
X-Cache-Status: MISS                          # first hit fetched backend; repeat for HIT.
```

## 3. Preflight (browser sends this before POST JSON)

```bash
curl -s -D- -o /dev/null -X OPTIONS localhost:8090/api/ping \
  -H 'Origin: http://app.test' \
  -H 'Access-Control-Request-Method: POST' \
  -H 'Access-Control-Request-Headers: Content-Type'
# 204 + Allow-Origin + Allow-Methods + Allow-Headers + Max-Age.
# Browser then sends the real POST. No preflight = no POST.
```

## 4. Same preflight without middleware → 404 (block by preflight)

```bash
curl -s -D- -o /dev/null -X OPTIONS localhost:8090/plain/ping \
  -H 'Origin: http://app.test' \
  -H 'Access-Control-Request-Method: POST'
# 404, no CORS headers. Blocked at the preflight gate: the real
# request is never sent. Contrast GET /plain (gate 2: 200 returned,
# then hidden from JS for missing ACAO).
```

## Rules distilled

- Simple requests (GET/POST form-encoded, no custom headers): 1 round trip.
- Non-simple (`Content-Type: application/json`, `Authorization`, `Prefer`): preflight first.
- Preflight answers `Access-Control-Allow-*`; browser enforces, server only declares.
- `*` origin + credentials never mix: echo the origin instead (ironclad rule, browsers reject `*` outright).

## Exposed vs hidden response headers (`GET /api/trace`)

```bash
curl -s -D- -o /dev/null localhost:8090/api/trace -H 'Origin: http://x.test' | grep -iE 'expose|x-'
# Access-Control-Expose-Headers: X-Request-Id
# X-Request-Id: req-123
# X-Internal-Flag: secret   (on the wire, invisible to JS)
```

- Frontend button reads both: `X-Request-Id` → value, `X-Internal-Flag` → `null`.
- Rule: JS sees safelisted headers + exposed ones only. Everything else exists on the wire but hidden from frontend code.

## Hand-rolled vs gin-contrib/cors (`cmd/gin-demo`, `:8092`)

Same routes, library middleware, mirrored config. Verified difference
on preflight:

```bash
make run-gin-demo   # terminal (:8092)
curl -s -D- -o /dev/null -X OPTIONS localhost:8092/api/ping \
  -H 'Origin: http://localhost:8091' \
  -H 'Access-Control-Request-Method: PUT' | grep -iE 'vary|max-age'
# Vary: Origin
# Vary: Access-Control-Request-Method
# Vary: Access-Control-Request-Headers
```

- Library varies on Origin **+ Method + Headers**: different preflights
  never share entries. Hand-rolled sends `Vary: Origin` only — same
  method/headers assumed, wrong when they differ.
- Lesson: libraries encode edge cases (Vary triple, credentials echo,
  origin validation) the hand version skips. Learn on hand-rolled,
  ship with the library.
- `POST /cred/login` + `GET /cred/me` exist on both servers via the
  shared `internal/session` package (HttpOnly cookie flow).
