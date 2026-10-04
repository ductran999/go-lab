# Minimal MCP — handshake, frames, tool envelope

> TL;DR: stdio carries newline-delimited JSON-RPC 2.0.
> `initialize` negotiates the version, `tools/list` advertises,
> `tools/call` runs. Tool failures ride home as `isError`
> results, never as protocol errors.

## 1. Frames

- One JSON object per line on stdin, one per line on stdout.
- `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` → response
  carries the same `id`. No `id` means notification (e.g.
  `notifications/initialized`) → server stays silent.
- Logs go to stderr. One stray log line on stdout desyncs
  every client — this is the classic stdio bug.

## 2. Handshake

| Direction                            | Shape                                                                    |
| ------------------------------------ | ------------------------------------------------------------------------ |
| client → `initialize`                | `{protocolVersion, capabilities, clientInfo}`                            |
| server → result                      | same `protocolVersion` echoed, `{capabilities:{tools:{}}}`, `serverInfo` |
| client → `notifications/initialized` | silence                                                                  |

Echoing the client's version is deliberate: hosts reject
servers that won't speak theirs.

## 3. Tool envelope

- `tools/list` → `{tools:[{name, description, inputSchema}]}` —
  the schema is the contract; the host builds validation and
  the agent's parameter help from it.
- `tools/call` `{name, arguments}` → `{content:[{type:text,
text}]}`. A failed tool still returns a result, with
  `isError:true` — protocol errors (`-32601` etc.) are only
  for "no such method/tool" and malformed frames.
- Args arrive `map[string]any` (numbers are float64): clamp
  and default server-side — the agent will send `n: 999999`.

## 4. Why stdlib, no SDK

- The surface we need is 4 methods and 3 tools — an SDK buys
  version negotiation and type safety we don't use yet.
- Zero new modules: this sandbox can't fetch any, and the lab
  stays `go vet` + `go test` clean offline.
- Graduate to the official SDK when the server needs sampling,
  roots, elicitation, or Streamable HTTP — not before.
