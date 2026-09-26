# Range lab — fetch slices, resume downloads

API on `:8106`. One readable ~1.63MB text (`line-000123 ...`, any
slice eyeball-verifiable). Full, partial, open-ended, past-the-end.

## Run

```bash
make run-server  # API on :8106
```

## Try

```bash
# Capabilities first
curl -s -D- -o /dev/null localhost:8106/file | grep -iE 'accept-ranges|etag|content-length'
# Accept-Ranges: bytes, ETag: "lines-v1", Content-Length ~1.63MB

# First 2 lines only (readable now)
curl -s localhost:8106/file -H 'Range: bytes=0-111'
# line-000000 the quick brown fox jumps over the lazy dog
# line-000001 ...

# Resume from byte 100 to the end (download resume shape)
curl -s localhost:8106/file -H 'Range: bytes=100-' -o ./tmp/tail.txt -w '%{http_code} %{size_download}\n'
# 206, total-100 bytes

# Past the end → 416 with the valid size hint
SIZE=$(curl -s localhost:8106/size | grep -o '[0-9]*')
curl -s -D- -o /dev/null localhost:8106/file -H "Range: bytes=${SIZE}-" | grep -iE 'HTTP|content-range'
# 416, Content-Range: bytes */SIZE

# curl's own resume flag speaks this protocol
curl -s localhost:8106/file -o ./tmp/full.txt -w '%{http_code}\n'
truncate -s 1000 ./tmp/full.txt
curl -s -C - localhost:8106/file -o ./tmp/full.txt -w '%{http_code}\n'
# 206, file whole again

# Parallel client: 5 ranges at once, merged locally, checksum-verified
make fetch
# ...merged... + "checksum ok" (Content-Digest header vs streamed
# sha256 of the merge — no second API, digest rides along)
# then: head -2 ./tmp/lines.txt && tail -1 ./tmp/lines.txt

# Error demo: big file + kill mid-download = corrupt, loud failure
FILE_MB=64 make run-server  # restart, terminal 1
./bin/fetch -parts 5 & sleep 0.3; pkill -x server; wait
# → fetch: ...connection reset... exit 1. Partial ./tmp/lines.txt
# left behind: ranges never promise atomicity, callers verify
# (checksum) before trusting the merge.
```

## Docs

- `docs/01-range.md` — 206/416, resume, seek
