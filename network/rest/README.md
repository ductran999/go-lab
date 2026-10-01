# REST versions — same dialect, faster vehicles

H1 (implicit everywhere), H2 (`http2/`), H3 (`http3/`): plain REST
JSON over three transports. Header mechanics live in `../headers/`.

| Dir | Transport |
|-----|-----------|
| `http2/` | H2C + TLS twins (prior knowledge, ALPN) |
| `http3/` | QUIC (needs `../rpc/grpc/certs/` demo CA) |
