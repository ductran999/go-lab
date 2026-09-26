# LB algorithms — when each wins

> TL;DR: stateless + even → **round-robin**. Uneven cost →
> **least-conn/latency**. Sticky sessions → **IP hash**.
> Known capacities → **weighted**. Real load signals → **resource-based**.

Run any: `go run ./network/load-balancing/cmd -app-name <flag>`.

| Flag | Algorithm | Wins when | Dies when |
|------|-----------|-----------|-----------|
| `rr` | Round-robin | Backends equal, requests even | Slow backend gets equal share (head-of-line by capacity) |
| `wr` | Weighted RR | Known static capacities (2x box = 2x weight) | Weights rot (capacity changes, nobody re-tunes) |
| `ih` | Source IP hash | Sticky state (carts, local cache) without shared store | Resharding reshuffles everyone; hot keys skew one box |
| `lc` | Least connections | Uneven request cost (long vs quick) | Counts lie with handoffs (streaming holds conns forever) |
| `ll` | Lowest latency | Backends degrade unevenly (noisy neighbor) | Latency lags (picks yesterday's winner, herd effect) |
| `rb` | Resource-based | Real CPU/mem signals available | Metrics stale/noisy → flapping between backends |

## Rules

- Default `rr`; upgrade on measured pain, not imagined scale.
- Sticky (`ih`) is a crutch for missing shared state — prefer
  external sessions, keep sticky for legacy only.
- Adaptive (`lc`/`ll`/`rb`) needs honest signals: active-request
  counters beat sampled metrics; damp changes (hysteresis) or
  the herd stampedes.
- Health checks sit underneath all six: no algorithm survives
  sending traffic to a dead box. Eject fast, probe back slowly.
- Ties to `../http/forwarding/`: the LB picks *where*, the trust
  policy decides *who* — gateways embed both.
