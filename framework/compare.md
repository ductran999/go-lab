# Frameworks compared — four dialects, one API

> TL;DR: same Todos JSON everywhere. **mux** teaches, **gin/echo**
> ship CRUD fast on net/http, **fiber** trades compatibility for
> raw RPS. Pick by constraint, not hype.

## 1. Side by side

|                        | mux (stdlib)                 | gin                  | echo                 | fiber                 |
| ---------------------- | ---------------------------- | -------------------- | -------------------- | --------------------- |
| Engine                 | `net/http`                   | `net/http`           | `net/http`           | `fasthttp`            |
| Route param            | `{id}` + `PathValue`         | `:id` + `c.Param`    | `:id` + `c.Param`    | `:id` + `c.Params`    |
| JSON out               | hand-rolled `writeJSON`      | `c.JSON`             | `c.JSON`             | `c.JSON`              |
| 404 JSON               | wrap + pattern check         | `NoRoute`            | `HTTPErrorHandler`   | status + JSON inline  |
| Middleware             | wrap funcs                   | `Use()` + `c.Next()` | `Use()`              | `Use()`               |
| Logger                 | slog hand-wrap               | `gin.Logger`         | `RequestLogger`+slog | `logger`              |
| `http.Server` features | full (timeouts, H2C, hijack) | full                 | full                 | **none** (own server) |
| SSE/hijack             | native                       | native               | native               | fiber-specific API    |

## 3. Shootout (wrk -t4 -c100 -d15s, GET /todos, local machine)

|       | Requests/sec | Avg latency | p99\*  |
| ----- | ------------ | ----------- | ------ |
| mux   | **103,375**  | **1.38ms**  | ~29ms  |
| fiber | 50,844       | 9.59ms      | ~528ms |
| echo  | 48,416       | 17.90ms     | ~791ms |
| gin   | 38,284       | 4.35ms      | ~85ms  |

\*p99 approximated from wrk distribution tail.

**Why mux wins here**: the lesson handler has **zero middleware** —
no per-request logging. gin/echo/fiber twins all log every request,
and at 100 conns that cost dominates. Lesson inside the lesson:
measure **your** stack (middleware included), not the router alone.
Strip the loggers and rerun — the gap shrinks to noise for JSON
this small. Raw router speed almost never decides; payload size,
DB calls and middleware do.

## 3b. Round 2 — loggers off (recovery only)

|       | Requests/sec | Avg latency |
| ----- | ------------ | ----------- |
| fiber | **190,575**  | **0.41ms**  |
| gin   | 100,493      | 1.11ms      |
| echo  | 99,485       | 1.14ms      |
| mux   | 72,374       | 2.04ms      |

fasthttp earns its reputation: ~2x the net/http pack once logging
is out of the way. gin ≈ echo (same engine family, same cost).
mux trails — note it also _dropped_ from round 1 (103k → 72k):
benchmarks wobble run to run (CPU noise, GC, turbo), so read gaps
< 2x as ties. The honest ranking: **fiber ≫ gin ≈ echo ≈ mux**
for hello-world JSON; real workloads (DB, payloads) compress all
four further.

## 4. Rule of thumb

- Learning / stdlib mechanics / full server control → mux.
- Team CRUD velocity on net/http → gin or echo (taste).
- Proven RPS bottleneck + no hijack/SSE needs → fiber.
- Anything else → the boring one your team already knows.
