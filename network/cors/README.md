# CORS Lab

One gin server, two behaviors: `/api/*` answers CORS preflights
(hand-rolled middleware, every header visible), `/plain/*` does not.

Deep-dive notes in [`docs/`](docs/): preflight matrix, credentials,
cache + Vary, browser enforcement. One doc per lab.

```bash
# Change DIR
$ cd network/cors
```

## Setup

```bash
make run-server   # API on :8090 (PORT to override)
make run-web      # demo page on :8091 (cross-origin vs :8090)
make cache-up     # nginx cache on :8085 (needs run-server)
make cache-down   # stop nginx cache
```

## Cleanup

No volumes, no state: stop the processes (`Ctrl+C`) and `make cache-down`.
