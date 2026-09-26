# PNA — public pages knocking on local devices

> TL;DR: Chrome treats public → private (local/LAN) as suspect:
> a **preflight** asks first, server opts in with
> **`Allow-Private-Network: true`**. No opt-in → handshake dies
> before 101. Origin allowlist rides alongside via `ORIGINS`.

## 1. Why it exists

- Attack shape: victim visits `https://evil.com`, its JS opens
  `ws://192.168.1.1/...` (your router) or `ws://localhost:APP`
  (your dev server). Same `CheckOrigin` story, enforced one
  layer earlier by the browser itself.
- Address spaces: public → local/private = private network
  request. local → local (our lab) is unaffected — PNA bites
  when the page is public and the socket is home.

## 2. Wire shape

```text
OPTIONS /ws
Origin: https://public.app
Access-Control-Request-Private-Network: true
→ 204 + Access-Control-Allow-Private-Network: true
THEN the normal WS handshake (Origin still checked).
```

- Two gates in series: PNA preflight (browser-enforced) then
  `CheckOrigin` (server-enforced). Either says no → no socket.
- curl simulates the preflight (see README); a true end-to-end
  needs a public page, out of lab scope.

## 3. Dynamic origins

- Hardcoded `localhost:8093` died: `ORIGINS` env holds the
  comma-separated allowlist (per-env config, no rebuild).
- Edge still strips/forwards honestly (see `../../http/forwarding/`);
  PNA + allowlist + gateway trust compose, each checks one thing.
