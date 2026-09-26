# REST on H2 — the transparent upgrade

> TL;DR: same routes, same JSON, same curl flags — only
> `proto=HTTP/2.0` in the log. Upgrading H1→H2 changes the
> **vehicle**, never the **dialect**.

## What changes

- Nothing in handlers: `http.ServeMux`, JSON bodies, status codes,
  middleware (auth, CORS, logging) work byte-identical.
- One line swaps the engine: `h2c.NewHandler(mux, &http2.Server{})`
  (cleartext) or `http.Server` + TLS + ALPN (browser path).

## What you gain for free

- Multiplexing (no 6-conn queue), HPACK (cheap repeat headers),
  stream cancel (`RST_STREAM` instead of killing connections).
- Single big download: same speed. Fifty small API calls: night
  and day — exactly the chatter REST lives on.

## Rules

- h2c is lab-only: browsers demand TLS (ALPN). Terminate TLS at
  the edge (gateway/LB), speak h2c behind it.
- Prior knowledge (client speaks H2 first byte) skips negotiation
  — fine inside a cluster, never across the internet (use ALPN).
