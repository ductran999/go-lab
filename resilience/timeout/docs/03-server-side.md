# Server-side timeouts — against dirty clients

Clients can also hang the server: every held connection is a
worker that never returns. (Our labs set `ReadHeaderTimeout: 5s`
for exactly this.)

| Timeout               | Stops                                                                  |
| --------------------- | ---------------------------------------------------------------------- |
| `ReadHeaderTimeout`   | Slowloris — header drip holding connections open                       |
| `ReadTimeout`         | slow body drip (giant POST arriving byte by byte)                      |
| `WriteTimeout`        | slow reader — handler done but client sips the response                |
| `IdleTimeout`         | clinically-dead keep-alive hogging a slot                              |
| `http.TimeoutHandler` | stuck handler (hung DB) — 503 after X, ctx cancelled                   |
| shutdown drain (30s)  | deploy killing in-flight — SIGTERM stops new, finishes old, then kills |

🪙 `WriteTimeout` kills SSE/streaming (long-lived by design) —
exempt those routes or scope the timeout per handler, never global.
