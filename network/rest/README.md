# REST versions — same dialect, faster vehicles

H1 (implicit everywhere), H2 (`http2/`), H3 (`http3/`): plain REST
JSON over three transports. Header mechanics live in `../headers/`.

| Dir      | Transport                                 |
| -------- | ----------------------------------------- |
| `http2/` | H2C + TLS twins (prior knowledge, ALPN)   |
| `http3/` | QUIC (needs `../rpc/grpc/certs/` demo CA) |

## Evolution: H1 → H2 → H3 (what each fixed)

|              | H1                               | H2                                                | H3 (QUIC/UDP)                  |
| ------------ | -------------------------------- | ------------------------------------------------- | ------------------------------ |
| Connections  | 6 TCP/origin (12 sockets)        | 1 TCP, N streams                                  | 1 UDP flow, N streams          |
| Framing      | Glued text                       | Binary frames + stream IDs                        | Same, over QUIC                |
| Head-of-line | Per-connection (req1 blocks all) | Per-connection (1 lost packet stalls all streams) | **Per-stream** (loss isolates) |
| Handshake    | TCP + TLS ≈ 3 RTT                | TCP + TLS ≈ 2 RTT                                 | **1 RTT (0 repeat)**           |
| Roaming      | Reconnect                        | Reconnect                                         | **Migration keeps it**         |
| Keep-alive   | Header needed                    | PING frames                                       | PING frames                    |
| Fixed        | —                                | 6-conn cap, header cost                           | TCP HoL, handshake, roaming    |

- Each generation keeps HTTP semantics (methods, status, headers)
  and fixes the transport pain of the last: H2 multiplexes, H3
  de-TCP-ifies. Same REST handlers throughout (`http2/` proves it).
- H3 wins where networks are bad (mobile, lossy); datacenters
  barely notice. H2 remains the lingua franca.
