# Content negotiation — client asks, server picks

> TL;DR: `Accept` ranks wanted shapes, server serves the **best
> match**, `Vary: Accept` keeps caches honest, miss → **406**.

## 1. Mechanics

```text
Accept: application/json;q=0.5, text/csv;q=0.9
```

- Comma list with optional `q=` weights (default 1). Server
  scores supported types, serves the winner with matching
  `Content-Type`.
- No header = no preference → serve the default (JSON here).
  Header present but nothing matches → `406 Not Acceptable`
  with the supported list (debuggable beats silent JSON).
- `Accept` (want) vs `Content-Type` (is): request `Content-Type`
  describes the **body sent** (POST/PUT); response `Content-Type`
  describes the **body returned**. Different jobs, same name.

## 2. Vary is load-bearing

- Cache key = URL + `Vary` headers. Without `Vary: Accept`,
  a cached CSV poisons the next JSON client at the same URL.
- Same class of bug as `Vary: Origin` in the CORS lab —
  under-varying serves wrong variants.

## 3. Where it shows up

- PostgREST: `Accept: text/csv` dumps tables, `Prefer` tunes
  insert/update shape — negotiation doing real work.
- Versioning via `Accept: application/vnd.api.v2+json`
  (content-type versioning) vs URL `/v2/` — header keeps URLs
  stable, URL keeps debugging easy. Pick one per API.
- Browsers send long Accept lists (`text/html, image/avif...
`); `q=` is how servers prefer WebP/AVIF over JPEG.
