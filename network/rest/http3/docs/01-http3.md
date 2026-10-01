# HTTP/3 in depth — QUIC: TCP+TLS+HTTP/2, redesigned as one

> TL;DR: QUIC merges handshake + encryption + multiplexing into
> **UDP**. Result: **0-RTT repeats**, **no TCP head-of-line**,
> **connection migration**. Same HTTP semantics — third plumbing.

## 1. Handshake (fewer trips)

```text
H1+TLS:  TCP (1 RTT) + TLS 1.2 (2 RTT) = 3 RTT before data
H2+TLS:  TCP (1 RTT) + TLS 1.3 (1 RTT) = 2 RTT
H3:      QUIC Initial + Handshake      = 1 RTT (0 on repeat!)
```
RTT = Round-Trip Time: one packet round trip (e.g. ~10ms HN→SG).
Every generation races to fewer round trips before first data.

- QUIC Initial packets carry the TLS 1.3 ClientHello **inside**
  (no separate TCP + TLS phases). Server replies with Handshake
  packets; then 1-RTT keys protect application data.
- **0-RTT resumption**: returning client sends data with the first
  flight (PSK from before). Fast — and replayable by design, so
  0-RTT requests must be idempotent (GET) or guarded server-side.

## 2. Streams without TCP

- QUIC streams ≈ H2 streams (IDs, flow control, FIN/RST) but each
  streams' bytes flow **independently**: one lost UDP packet stalls
  only its stream, not the connection. The last head-of-line dies.
- Connection = **connection ID**, not 4-tuple. IP/port changes
  (wifi→5G) keep the connection alive: **migration**, the mobile
  AI-chat superpower from the earlier discussion.

## 3. H2 vs H3

| | H2 (TCP) | H3 (QUIC/UDP) |
|---|---|---|
| Handshake | 2 RTT (TCP+TLS) | 1 RTT (0 repeat) |
| Lost packet | Stalls all streams | Stalls one stream |
| Roaming | New connection | Migration keeps it |
| Middleboxes | TCP understood everywhere | UDP throttled/blocked some nets |
| Debugging | Wireshark reads TCP reassembly | Encrypted packets (keys needed) |

## 4. Rules

- TLS is **mandatory** (no h2c equivalent): QUIC packets are
  always encrypted past Initial. Our lab reuses the grpc demo CA.
- Same HTTP: methods, status, headers, SSE, `traceparent` —
  unchanged. Only `r.Proto` flips to `HTTP/3.0`.
- Deploy note: UDP firewalls + LB support lag TCP. H3 fastest
  where the network is worst (mobile, lossy) — datacenters barely
  notice.
