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

## Headers vs body: split frames, same stream

```text
H1 (glued):   GET /todos HTTP/1.1\r\nHost: ...\r\n\r\n{"task":"x"}
              └──────── one text blob, read in order ────────┘

H2 (split, paired by stream ID):
  stream 1: HEADERS(method, auth...) → DATA("hel") → DATA("lo", END)
  stream 3: HEADERS(...)              → DATA(...)  (interleaved freely)
              └──── same ID = same pair, never mixed ────┘
```

- H1: headers+body glued text — read all headers before body,
  cancel kills the connection.
- H2: HEADERS frames (HPACK) + DATA frames travel separately and
  interleave; server may respond before reading the full body,
  and RST_STREAM cuts mid-body cleanly.
- One-liner: H2 labels every frame with a stream ID so both sides
  reassemble full semantics in any byte order; H1's glued blocks
  mean req1 unfinished blocks the whole TCP (head-of-line).

## Socket economics (the whole H2 pitch in numbers)

- **H1**: 6 parallel TCP/origin = **12 sockets** (6 client + 6 server)
  for 6 in-flight requests. Request 7 queues.
- **H2**: 1 TCP (2 sockets), **N streams** — the 6-cap is gone
  (remaining cap: `max_concurrent_streams`, 100+ by default).
- 12 sockets for 6 jobs → 2 sockets for N jobs. Auth, routes and
  JSON unchanged: each stream carries its own headers, verified
  independently like H1.

## Rules

- h2c is lab-only: browsers demand TLS (ALPN). Terminate TLS at
  the edge (gateway/LB), speak h2c behind it.
- Prior knowledge (client speaks H2 first byte) skips negotiation
  — fine inside a cluster, never across the internet (use ALPN).
