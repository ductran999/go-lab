# Auth headers — where tokens ride

> TL;DR: **header** for APIs, **HttpOnly cookie** for browser
> sessions, **query** only when headers are impossible (SSE,
> WS handshake, links). Server **verifies signature**, never
> trusts shape.

## 1. Bearer anatomy

```text
Authorization: Bearer <header>.<payload>.<signature>
```

- Three base64url parts. Header + payload are **readable**
  (decode, don't trust). Signature (HMAC/RSA over the first
  two) is the only trust root — verify it, check `exp`.
- JWT ≠ encryption. Secrets in payload leak to anyone holding
  the token (DevTools, logs, proxies).

## 2. Three channels

|                          | Header                               | Cookie (HttpOnly)         | Query                  |
| ------------------------ | ------------------------------------ | ------------------------- | ---------------------- |
| JS-readable              | Yes (XSS steals)                     | No                        | Yes (worst)            |
| CSRF                     | Immune (not auto-sent)               | Needs SameSite/CSRF token | Immune                 |
| CORS preflight           | Yes (`Authorization` non-safelisted) | With creds mode           | No                     |
| Usable by EventSource/WS | No (no custom headers)               | Yes (auto-sent)           | Yes                    |
| Leaks into               | —                                    | —                         | logs, history, Referer |

- **Header**: default for API clients (curl, services, SPAs
  holding tokens in memory). Preflight cost per origin is
  one OPTIONS, cached via `Access-Control-Max-Age`.
- **Cookie**: browser sessions. `HttpOnly` beats XSS theft;
  `SameSite` beats CSRF. Cross-site needs `None; Secure` and
  increasingly fights third-party-cookie blocking — the web
  is pushing sessions back to same-origin + BFF.
- **Query**: escape hatch for channels without headers
  (EventSource, WS handshake, `<img>`, magic links).
  Short-lived, single-purpose, logged everywhere — treat as
  semi-public.

## 3. Ties to earlier labs

- SSE (`../sse/`): no headers → query or cookie, short-lived.
- WS (`../websocket/`): handshake headers settable only by
  non-browser clients; browser WS auth = cookie/query, or
  first-message token verified before subscribing.
- CORS (`../cors/`): `Allow-Headers` must list `Authorization`
  or preflight fails; credentials mode needs echoed origin +
  `Allow-Credentials` (see `02-credentials.md`).

## 4. Server rules

- Accept **one** channel per endpoint (`/me-strict` pattern);
  multi-channel (`/me`) is a demo, not a policy.
- 401 on missing/malformed/expired — same status, no oracle
  distinguishing which check failed (ours leaks detail for
  teaching; production says `invalid or missing token` only).
- Secrets ≥ 32 chars, from env, never in repo (gitleaks watches).
