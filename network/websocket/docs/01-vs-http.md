# WebSocket vs HTTP/1.1

**TL;DR:** HTTP/1.1 = request/response per message (stateless, headers each time). WebSocket = one handshake, then full-duplex frames both ways (stateful, ~2 bytes overhead).

Keywords: handshake, upgrade, frames, full-duplex, stateful.

## Handshake (the only HTTP part)

```text
Client → GET /ws + Upgrade: websocket + Sec-WebSocket-Key
Server → 101 Switching Protocols (+ Sec-WebSocket-Accept)
After:  HTTP done. Raw TCP + frames from here.
```

- Upgrade reuses HTTP infra (ports, TLS, proxies) to bootstrap.
- After 101, no methods/paths/statuses — just frames.

## Message flow compared

|             | HTTP/1.1                               | WebSocket                                      |
| ----------- | -------------------------------------- | ---------------------------------------------- |
| Pattern     | req → resp, half-duplex                | frames either way, full-duplex                 |
| Headers     | Every message (~500B+)                 | Once (handshake), then ~2B/frame               |
| State       | Stateless (cookies/sessions bolted on) | Stateful socket (server knows who's connected) |
| Server push | No (poll/long-poll/SSE workarounds)    | Native both directions                         |
| Direction   | Client always initiates                | Either side anytime                            |
| Closing     | Per-response / keep-alive reuse        | Close frames, ping/pong keepalive              |

## vs SSE (sibling lab)

- SSE: server→client only, HTTP-based (reconnect + `Last-Event-ID` free), text only.
- WS: both directions, binary capable, needs manual heartbeat/reconnect.
- Rule: display-only live feed → SSE. Chat/game/collab (client talks back) → WebSocket.

## Costs of going stateful

- 1 socket per client on the server (memory + FDs, sticky sessions behind LBs).
- Proxies/LBs must support upgrade (some buffer or kill idle sockets).
- Reconnect, heartbeat, backoff: app-level work HTTP gives for free.
