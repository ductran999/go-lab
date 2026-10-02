# Browser Enforcement

**TL;DR:** CORS blocks at the **browser**, not the server: bytes arrive,
JS can't read them. Server declares (headers + preflight answers),
browser enforces (hides mismatches from JS). Same URL: curl always
200, browser may block. Fire-and-forget (`<img>`, form POST, `no-cors`)
still sends — the server still executes; only the *read* is gated.

Keywords: same-origin policy, gates, TypeError.

```bash
make run-server   # terminal 1 (:8090)
make run-web      # terminal 2 (:8091)
# open http://localhost:8091, DevTools console open
```

- `GET /api/ping` → 200 shown (ACAO allows it).
- `GET /plain/ping` → BLOCKED TypeError (server still 200 in Network tab,
  browser hides it from JS). curl to the same URL always succeeds.
- `PUT /api/ping` → OPTIONS preflight first (Network tab), then real
  request. Toggle preflight cache (Disable cache) to see it every time.

## Two gates

- Gate 1, preflight: OPTIONS fails (404, method disallowed) → real request never sent.
- Gate 2, read: 200 returned → hidden from JS for missing/mismatched ACAO.
- Same root cause for `/plain`: no CORS configured. Fix one place (middleware), both gates open.

## Exposed vs hidden response headers (`GET /api/trace`)

```bash
curl -s -D- -o /dev/null localhost:8090/api/trace -H 'Origin: http://x.test' | grep -iE 'expose|x-'
# Access-Control-Expose-Headers: X-Request-Id
# X-Request-Id: req-123
# X-Internal-Flag: secret   (on the wire, invisible to JS)
```

- Frontend reads both: `X-Request-Id` → value, `X-Internal-Flag` → `null`.
- Rule: JS sees safelisted + exposed headers only. Expose answers
  *what JS may read* — the read gate, independent of preflight
  (which answers *whether JS may send*).
