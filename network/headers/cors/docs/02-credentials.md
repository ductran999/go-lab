# Credentials Mode

**TL;DR:** Credentials need `Allow-Credentials: true` + echoed Origin. `*` is rejected outright (ironclad rule).

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
