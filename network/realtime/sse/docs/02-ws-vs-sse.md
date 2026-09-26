# WS vs SSE — pick by direction

TL;DR: client talks back → WebSocket. Server-only push →
SSE (simpler, auto-resume, plain HTTP tooling works).

## 1. Head-to-head

|                 | WebSocket (`../websocket/`)    | SSE (this lab)                 |
| --------------- | ------------------------------ | ------------------------------ |
| Direction       | Full-duplex                    | Server → client only           |
| Handshake       | HTTP Upgrade → frames          | Plain GET, stays open          |
| Client → server | Frames anytime                 | Separate POST/fetch            |
| Reconnect       | Manual (your code)             | Built-in + `retry:`            |
| Resume          | Manual (your offsets)          | `Last-Event-ID`                |
| Data            | Text + binary                  | Text (`data:` lines)           |
| Browser API     | `WebSocket`                    | `EventSource`                  |
| Same-origin     | Server decides (`CheckOrigin`) | CORS headers decide            |
| Masking         | Client frames masked           | None (plain HTTP)              |
| Proxies/CDN     | Needs WS-aware hops            | Works, but mind buffering      |
| Debugging       | Frames tab                     | Response stream in Network tab |

## 2. Decision rule

- Chat, games, collaborative editing (both sides chatty) → WS.
- Feeds, notifications, progress bars, live dashboards → SSE.
- Realtime fan-out to thousands of browsers → SSE (HTTP infra:
  load balancers, compression, auth middleware all apply unchanged).
- Binary blobs both ways → WS.

## 3. Why our realtime lab chose SSE

`storage/realtime/` bridges Postgres NOTIFY to `/stream`: the flow
is inherently one-way (DB → browser), many listeners, resume-friendly.
SSE gives reconnect + `Last-Event-ID` for free; WS would add framing,
masking, and manual resume code for zero benefit.
