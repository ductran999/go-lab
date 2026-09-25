# HTTP/2 in depth — one connection, many conversations

> TL;DR: H2 keeps HTTP semantics, replaces the text pipe with
> **binary frames** grouped into **streams**. Result: **multiplexing**,
> **header compression**, **flow control** — one TCP connection does
> the work of six.

## 1. Setup: how H2 starts

- **HTTPS (the norm)**: TLS handshake with **ALPN** — client lists
  `h2,http/1.1`, server picks. Upgrade negotiated before first byte
  of HTTP. Browsers only do H2 over TLS.
- **h2c (cleartext)**: HTTP/1.1 request with `Upgrade: h2c`, server
  answers 101 (same dance as WebSocket!) — or **prior knowledge**:
  client speaks H2 frames immediately (gRPC's `insecure` mode does
  this: no handshake dance at all).
- **Preface**: client sends magic `PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n`
  then a SETTINGS frame. Server replies SETTINGS. Tunables exchanged
  (max concurrent streams, window sizes, header table size).

## 2. Frames: the binary alphabet

Every frame: 9-byte header (`length | type | flags | stream id`) + payload.

| Type            | Job                                                        |
| --------------- | ---------------------------------------------------------- |
| `HEADERS`       | Opens a stream, carries HPACK'd headers (+ priority)       |
| `DATA`          | Body bytes                                                 |
| `SETTINGS`      | Connection knobs (both directions)                         |
| `WINDOW_UPDATE` | Flow-control credits (per stream + per connection)         |
| `PRIORITY`      | Stream weight/dependency (rarely honored today)            |
| `RST_STREAM`    | Abort one stream (cheap cancel — H1 can't do this)         |
| `GOAWAY`        | Graceful connection shutdown (finish streams, send no new) |
| `PING`          | RTT probe + keepalive                                      |
| `CONTINUATION`  | Header block overflow (many headers)                       |

- **Stream IDs**: odd = client-initiated, even = server-initiated
  (push), `0` = connection-level. IDs never reused on a connection.
- **States**: idle → open (HEADERS) → half-closed (one side END_STREAM)
  → closed. Either side `RST_STREAM`s anytime — canceling one download
  doesn't touch the other 31 sharing the pipe.

## 3. Multiplexing (the headline)

- H1: 1 connection = 1 request at a time → browsers open ~6.
- H2: N streams interleave frames on 1 connection. Stream A's slow
  DATA doesn't block stream B's HEADERS — different frames, same wire.
- **Flow control**: each stream + the connection hold byte windows;
  receiver grants more via `WINDOW_UPDATE`. A greedy download can't
  starve sibling streams; backpressure is protocol-native (compare
  our labs' hand-rolled drop-on-full Hub buffers).
- **Cost**: one lost TCP packet stalls _every_ stream (TCP reorders
  below H2's view). HTTP-over-QUIC (H3) fixes this with per-stream
  UDP flows — the remaining head-of-line moves down a layer again.

## 4. HPACK: headers that shrink

- H1 resends `user-agent`, `cookie`, `accept...` verbatim per request.
- HPACK: static table (common headers pre-numbered) + dynamic table
  (per-connection LRU of seen headers) + Huffman coding. Repeat
  `cookie: session=abc` costs an index byte, not 20.
- Why it matters for gRPC: chatty RPCs (auth metadata per call)
  stay cheap; H1 would re-send kilobytes of headers each round.

## 5. Server push (rise and fall)

- Server could push resources before asked (`PUSH_PROMISE` + stream).
  Dream: answer + its CSS/JS in one round trip.
- Reality: caches duplicate pushed bytes, browsers rarely used it
  well — **deprecated**, removed from Chrome. Lesson: the network
  can't outsmart the cache; `103 Early Hints` + preload won instead.

## 6. Labeled envelopes on one TCP (the multiplexing trick)

```text
→ HEADERS  stream=1  (GET /a)
→ HEADERS  stream=3  (GET /b)
→ DATA     stream=1  "hel"
→ DATA     stream=3  "img-b..."
← DATA     stream=3  "done-b" END
→ DATA     stream=1  "lo" END
← DATA     stream=1  "done-a" END
```

- Bytes of both streams interleave freely; the receiver sorts by
  **stream ID** back into request/response pairs. H1 sees this
  byte salad as garbage — no IDs, no lengths, no way to split.
- No keep-alive header needed: H2 connections are long-lived by
  design, with **PING** frames for probing and **GOAWAY** for
  graceful shutdown. Keep-alive is H1 heritage, retired here.

## 7. Where our labs sit

- gRPC: H2-native (needs streams). Reflection, health checks ride
  the same connection as RPCs.
- SSE: works on both — plain response trick, H2 just multiplexes it
  with siblings (no more 6-slot math).
- WS: H1-only in browsers (RFC 8441 defines WS-over-H2; nobody
  ships it). Six-socket pool still applies.
- Everything above framing — methods, status, CORS, cookies, auth,
  `traceparent` — is byte-identical on H1 and H2.
