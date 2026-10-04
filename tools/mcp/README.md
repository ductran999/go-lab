# MCP go-lab — labs as agent tools, stdlib only

Stdio MCP server: no SDK, just JSON-RPC 2.0 over stdin/stdout
(one object per line, logs on stderr). Run from the repo root —
`OTEL_LAB_DIR` defaults to `observability/otel`.

## Run

```bash
make run-server  # speaks MCP on stdio (nothing visible until you send frames)
```

## Try

```bash
# handshake + tool list
printf '%s\n' \
 '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}' \
 '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | go run ./cmd/server 2>/dev/null

# grep the live exposition (needs svc-a metrics on :2112)
printf '%s\n' \
 '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"metrics_query","arguments":{"pattern":"http_requests_total"}}}' \
 | go run ./cmd/server 2>/dev/null | head -c 600

# retry buckets (needs retry lab on :8122)
printf '%s\n' \
 '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"retry_stats"}}' \
 | go run ./cmd/server 2>/dev/null
```

## Tools

| Tool                      | Does                                                      | Needs                |
| ------------------------- | --------------------------------------------------------- | -------------------- |
| `bench_run` (n, c)        | fires `otel/bench.sh`, returns delta + buckets + exemplar | svc-a/svc-b up       |
| `metrics_query` (pattern) | greps `/metrics`, max 50 lines                            | metrics on `:2112`   |
| `retry_stats`             | dumps retry `/stats` buckets                              | retry lab on `:8122` |

## Wire into opencode

Add to `opencode.json` (absolute path, env pins the lab dirs):

```json
{
  "mcp": {
    "go-lab": {
      "type": "local",
      "command": ["go", "run", "./cmd/server"],
      "directory": "/home/danny/Study/side-project/go-lab/tools/mcp",
      "environment": {
        "OTEL_LAB_DIR": "/home/danny/Study/side-project/go-lab/observability/otel"
      }
    }
  }
}
```

## Docs

- `docs/01-mcp-minimal.md` — handshake, frames, tool envelope
