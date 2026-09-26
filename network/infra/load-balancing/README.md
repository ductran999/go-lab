# Load-balancing lab — six algorithms, one harness

Moved from the repo root (`load-balancing-alg/`): each algorithm
spins backends and balances across them. Pick with `-app-name`.

## Run

```bash
go run ./network/load-balancing/cmd -app-name rr  # from repo root
```

| Flag | Algorithm | Picks by |
|------|-----------|----------|
| `rr` | Round-robin | Turn order |
| `wr` | Weighted round-robin | Static weights |
| `ih` | Source IP hash | Client IP (sticky) |
| `lc` | Least connections | Fewest live requests |
| `ll` | Lowest latency | Fastest recent replies |
| `rb` | Resource-based | Backend load metrics |

## Docs

- `docs/01-algorithms.md` — when each wins, failure modes, rules
