# HTTP/3 lab — QUIC streams, TLS mandatory

Server on `:8108/udp` (QUIC, not TCP). Unary `/hello` + finite
`/stream`. Certs shared with the grpc lab demo CA (`localhost`).

```bash
# Change DIR (relative cert paths ../grpc/certs resolve from here)
$ cd network/http3
```

## Run

```bash
make run-server  # terminal 1 — :8108/udp
make run-client  # terminal 2 — hello + stream over QUIC
```

Wireshark: capture `lo`, filter `quic` — handshake (Initial),
HANDSHAKE packets (TLS inside), then 1-RTT protected payload.
Compare with the grpc lab's `http2` capture: same ideas, no TCP.

## Docs

- `docs/01-http3.md` — QUIC handshake, streams, migration, vs H2
