# Harness skeleton — the loop around the brain

Stdlib only. The loop, tool registry, and permission gate are the
real architecture; only the brain is mocked (scripted steps) since
model APIs need network. Swap `MockBrain` for an API call and the
rest stays identical.

## Run

```bash
go run ./cmd/demo
# [ask] permit metrics_query? → yes (scripted)
# answer: buckets spread, no restart needed, traffic nominal
```

The script tries `restart_service` mid-run: the gate denies
`restart_*`, the loop records `denied` and continues — the same
gate that asks you before my bash commands.

## Parts

| Piece      | Job                                                                  |
| ---------- | -------------------------------------------------------------------- |
| `Brain`    | picks next `Action` from history (mock replays a script)             |
| `Tools`    | `name → func`, the agent's hands (MCP servers plug in here)          |
| `Gate`     | allow/ask/deny per tool pattern; ask without `AskFunc` defaults deny |
| `MaxSteps` | budget — runaway loops die with `budget out`, not with your money    |
| history    | every act/obs/denial appended — the brain's entire memory            |
