# HTTP caching — freshness first, validation second

> TL;DR: **freshness** (`max-age`) avoids the trip, **validation**
> (`ETag` + `304`) saves the bytes, **SWR** hides the wait,
> **`Vary`** splits the key. Fresh > validated > stale-served > fetched.

## 1. Freshness (no trip at all)

```text
Cache-Control: public, max-age=60
```

- Response reusable for 60s **without contacting origin**.
  `/hits` stays flat — the strongest saving.
- `public` = shared caches (CDN/proxy) may store. `private` =
  browser only (per-user data). Default without either: shared
  caches refuse, browsers decide heuristically — be explicit.
- `no-cache` ≠ no storing: it means **revalidate every time**
  (misleading name, historic). `no-store` = never keep anything.

## 2. Validation (trip, no bytes)

```text
ETag: "1"  →  If-None-Match: "1"  →  304 Not Modified
```

- Stale entry revalidated: origin compares, answers **304 with
  empty body** when unchanged. Trip spent, bytes spared.
- Strong validators (byte-identical `ETag`) vs weak (`W/"1"`,
  semantically same). `Last-Modified`/`If-Modified-Since` is the
  older, second-precision cousin — ETag wins on precision.
- Our lab: 304s still bump `/hits`. Validation saves bandwidth,
  never latency.

## 3. Stale-while-revalidate (hide the wait)

```text
Cache-Control: max-age=10, stale-while-revalidate=50
```

- 0–10s: serve fresh. 10–60s: serve **stale instantly** +
  revalidate in background. User never waits past 10s.
- `stale-if-error=60` goes further: origin down → keep serving
  stale instead of failing. Resilience for free.
- Needs a real cache in front to observe (browser honors it;
  our cors nginx PoC at `:8085` is the next place to demo it).

## 4. Vary (split the key)

- Cache key = URL + selected request headers. `Vary: Origin`
  (cors lab `03-cache-vary.md`) keeps per-origin entries —
  without it, one origin's CORS headers poison another's.
- `Vary: Accept-Encoding` (gzip vs plain), `Vary: Accept-Language`.
  Over-varying fragments the cache (low hit rate); under-varying
  serves wrong variants. Vary exactly on what changes the body.

## 5. Rule of thumb

- Immutable assets (hashed filenames) → `max-age=31536000, immutable`.
- Per-user API data → `private, no-cache` + `ETag` (revalidate, spare bytes).
- Hot public data → short `max-age` + SWR (fast + fresh-ish).
- Auth/secret/personal → `no-store`, no debate.

## 6. Mental model (Q&A notes)

- **Origin labels, caches store.** Origin (`:8099`) only sets
  headers; browser (private), CDN/proxy (shared) keep the bytes.
  One declaration, every HTTP-aware hop obeys — no coordination.
- **Caches ask, origin answers.** Stale/miss → cache sends a
  plain `GET` with validators (`If-None-Match`); origin replies
  `304` (keep yours) or `200` (here is new).
- **Fresh = no request.** Within `max-age` the request dies at
  browser/proxy — never on the wire, origin blind, `/hits` flat.
- **Dynamic bodies must not be fresh.** Per-user/per-call data
  → `no-store`, or `private, no-cache` + content-based `ETag`.
  CRUD: ETag per resource version, mutations invalidate —
  never long `max-age` on lists.
- **The loop in one line:** server stamps `ETag`, client asks
  `If-None-Match`, match → `304`, else `200` with a new stamp.
