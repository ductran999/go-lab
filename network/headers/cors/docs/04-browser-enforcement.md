# Browser Enforcement

**TL;DR:** Server declares (headers + preflight answers), browser enforces (hides mismatches from JS). Same URL: curl always 200, browser may block.

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
