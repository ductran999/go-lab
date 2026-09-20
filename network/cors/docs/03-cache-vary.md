# Cache + Vary

**TL;DR:** No `Vary` → origins share one HIT (wrong variant). `Vary: Origin` → one entry per Origin. Verified live both directions.

Keywords: variants, poisoning, fingerprints.

Rule: no `Vary` → sites A and B share one HIT (wrong variant served);
with `Vary: Origin` → one entry per Origin (correct).

## PoC (`cache/`)

nginx (`:8085`) in front of the lab API, default cache key (no Origin):

```bash
make cache-up     # needs make run-server in another terminal
# A with Origin -> MISS (populates entry)
curl -s -D- -o /dev/null localhost:8085/api/ping -H 'Origin: http://a.test' | grep -iE 'X-Cache|allow-origin'
# B other Origin -> MISS again (separate entry, see below)
curl -s -D- -o /dev/null localhost:8085/api/ping -H 'Origin: http://b.test' | grep -iE 'X-Cache|allow-origin'
make cache-down
```

Observed both directions:

- **With `Vary: Origin`** (our `/api`): nginx splits entries per Origin. B = MISS. Correct.
- **Without `Vary`** (a no-Origin response cached first): that entry serves **all** origins — Origin requests got a variant with no ACAO headers. The bug, live.

Fixes when the backend omits `Vary`: key the cache on Origin explicitly
(`proxy_cache_key "...$http_origin"`) or make the backend send `Vary`.
Verified live, both directions.

## Reading a cached response (proxy fingerprints)

```bash
HTTP/1.1 200 OK
Server: nginx/1.31.6              # proxy signature (prod: server_tokens off)
Content-Type: application/json; charset=utf-8
Content-Length: 18
Connection: keep-alive            # hop-by-hop: browser<->proxy leg only
X-Cache-Status: HIT               # served from cache, backend untouched
```

- `Server` leaks software + version: recon fuel, hide it in prod.
- `Connection` never forwards past the proxy hop.
- `X-Cache-Status` is debug-only: strip publicly or keep internal.
