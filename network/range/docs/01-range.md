# Range — slices over HTTP

> TL;DR: `Accept-Ranges: bytes` advertises slicing, `Range`
> asks for a slice, **206** serves it with `Content-Range`,
> past-the-end → **416**. Resume and seek are the same trick.

## 1. Shapes

```text
Range: bytes=0-15       → 206, Content-Range: bytes 0-15/1709179   (closed)
Range: bytes=100-       → 206, Content-Range: bytes 100-1709178/…  (resume)
Range: bytes=-500       → 206, last 500 bytes                      (suffix)
Range: bytes=9999999-   → 416, Content-Range: bytes */1709179      (miss)
```

- `If-Range: "etag"` guards resume: file changed mid-download
  → server ignores Range, sends full 200 (stale tail never glued
  to a new head).
- Multi-range (`bytes=0-10,20-30`) exists (`multipart/byteranges`)
  — rarely used, most clients ask serially.

## 2. Uses

- **Resume**: `curl -C -` / download managers: HEAD size known,
  append from `bytes=<have>-`, glue locally.
- **Seek**: video players probe `Accept-Ranges`, then jump
  (`bytes=<position>-`) instead of downloading hours of video.
- **Parallel fetch**: split 1GB into N ranges, N connections,
  concatenate — poor-man's throughput (mind the 6-conn pool
  per origin from the SSE lab).

## 3. Rules

- Advertise `Accept-Ranges: bytes` only when slicing is real
  (static files, versioned blobs). Dynamic bodies → omit it.
- Pair with `ETag` so resume invalidates on change (If-Range).
- 416 must include `Content-Range: bytes */<size>` — the size
  hint is how clients clamp their next ask.
- Ranges are not atomic: parallel merges can die half-written
  (kill mid-download → corrupt tail). Always verify against the
  checksum traveling with the response (`Content-Digest`, RFC 9530)
  before trusting the file — stream the hash, never `ReadAll`
  a multi-GB merge. Sidecars (`.sha256` files, manifest digests)
  are the same idea offline.

## 5. Streaming hash (why chunked hashing equals whole hashing)

- Hash keeps one running **state** (32 bytes), not the bytes:
  `state = f(state, nextBytes)`. Cut boundaries never enter
  the math — `Write(a)+Write(b)` ≡ `Write(a+b)`.
- Digest = function of **byte order only**. Same order → same
  output, whether hashed whole, in 32KB loop buffers, or in
  awkward cuts (`cmd/hashproof` proves it: whole vs 7/65537/rest).
- `io.Copy(hasher, file)` reuses one 32KB buffer: each read
  **overwrites** the last, RAM stays flat for any file size.

## 4. Probe before slicing (HEAD anatomy)

```text
HEAD /file → 200 (never 206: no slice asked, nothing cut)
Accept-Ranges: bytes          <- slicing supported, safe to split
Content-Length: 1709179       <- total: plan splits + pre-size output
Content-Type: application/octet-stream  <- raw bytes, do not render
ETag: "blob-v1"               <- version: If-Range guards resume
Last-Modified / Date          <- fallback validators, auto-set
```

- Fetch clients (`cmd/fetch`) HEAD first: no ranges → abort,
  unknown size → abort. Never slice blind.
- `If-Range: "etag"` on resume: file changed mid-download →
  server ignores Range, sends full 200. Stale tails never glue
  onto new heads.
