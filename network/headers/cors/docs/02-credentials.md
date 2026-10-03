# Credentials Mode

**TL;DR:**

- **Send**: browser needs `credentials: 'include'` (default: nothing cross-origin).
- **Accept**: server needs echoed `Origin` (never `*`) + `Allow-Credentials: true` + `Vary: Origin`.
- **Modes**: Strict = never cross-site · Lax = links yes, fetch no · None+Secure = always (TLS only).

Keywords: cookies, echo, Vary.

```bash
curl -s -D- -o /dev/null localhost:8090/cred/ping -H 'Origin: http://app.test' | grep -iE 'allow-origin|credentials|vary'
# Access-Control-Allow-Origin: http://app.test  (echoed, never *)
# Access-Control-Allow-Credentials: true
# Vary: Origin  (load-bearing: each origin caches its own ACAO)
```

- Browser: `fetch(url, {credentials: 'include'})` sends cookies; mismatched ACAO blocks.
- Frontend button: `GET /cred/ping with credentials`. Same-origin page + `*` endpoint with creds mode also blocks — try it.

## Fetch credentials modes (request side)

- `omit` — send nothing (cross-origin default). Safest, no cookies involved.
- `same-origin` — send only same-origin (general default).
- `include` — always send, even cross-origin. Response then needs matching ACAO + `Allow-Credentials: true` for JS to read it.

Sending and reading are independent: cookies ride along (jar rules) whether or not JS may read the response.

## Headers don't hide

- DevTools/curl/proxies see **all** headers. `Expose-Headers` gates page-JS reads only.
- No secrets in custom headers. Ever.

## SameSite levels (needs 2 hostnames to demo, ports don't count)

Setup: `/etc/hosts` → `127.0.0.1 app.test api.test`. Page on `app.test:8091`, API on `api.test:8090` = cross-site.

| Cookie                   | Link click app→api (top-level GET) | fetch POST cross-site | Use when                           |
| ------------------------ | ---------------------------------- | --------------------- | ---------------------------------- |
| **Strict**               | ❌ Dropped (logged out)            | ❌                    | Banking, admin                     |
| **Lax** (modern default) | ✅ Sent (straight to dashboard)    | ❌                    | Regular web, balanced              |
| **None** + `Secure`      | ✅                                 | ✅                    | Cross-site embeds, third-party SSO |

Live demo (page MUST open via `http://app.test:8091`, API base switched to `api.test:8090`):

1. Click **Set Strict/Lax/None cookie** → `GET /cred/mode/:mode` (link target updates).
2. Click **fetch /cred/me** (background): Strict → 401 always; Lax → 401; None → 200.
3. Click **Open /cred/whoami** (top-level navigation): Strict → 401; Lax + None → `user: demo-user-1`.

Same cookie, two fates by context — that is the whole Lax lesson.

```text
browser tab
└── TOP document (app.test:8091)          ← top-level: address bar points here
    │   • link click / URL type / submit → whole page changes (navigation)
    │   • Lax cookie: SENT (user is driving)
    │
    ├── <iframe api.test>                 ← nested: page inside page
    │     Lax cookie: NOT sent (background)
    ├── fetch() ──→ api.test              ← background: JS behind your back
    │     Lax cookie: NOT sent
    └── <img src=api.test>                ← subresource
          Lax cookie: NOT sent
```

Top-level = the root browsing context (what the address bar shows).
Everything inside it (iframes, fetch, images) is nested/background:
Lax trusts page changes, suspects background work.

- Port differs ≠ different site. Same scheme + registrable domain = same site (`:8091` → `:8090` is same-site).
- Prod combo: `HttpOnly + Secure + Lax`. `localhost` counts as trustworthy, so `Secure` works over plain http locally.

## Cookie proof (`POST /cred/login` → `GET /cred/me`)

```bash
curl -c jar -b jar -s -X POST localhost:8090/cred/login | head -c 60; echo
# login sets: session=demo-user-1; HttpOnly (JS can't read it, check DevTools)
curl -s localhost:8090/cred/me -b jar
# {"user":"demo-user-1"} — server received the cookie
curl -s -o /dev/null -w '%{http_code}\n' localhost:8090/cred/me
# 401 — no cookie, no user. Header or no header, backend validates alone.
```

- `Vary: Origin` is mandatory here, not optional: ACAO differs per origin, sharing entries leaks one origin's permission into another's.
