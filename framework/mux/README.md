# mux deep-dive — stdlib `net/http.ServeMux`, no magic (not gorilla/mux)

One day on `net/http`: routing, middleware, lifecycle, context,
static, testing. Each lesson runs standalone.

```bash
# Change DIR
$ cd network/http/mux
```

## Lessons

| #   | Dir                      | What                                     |
| --- | ------------------------ | ---------------------------------------- |
| 1   | `lessons/01-routing/`    | Patterns, methods, wildcards, precedence |
| 2   | `lessons/02-middleware/` | Wrap, chain order, per-route vs global   |
| 3   | `lessons/03-lifecycle/`  | Timeouts, graceful shutdown              |
| 4   | `lessons/04-context/`    | Request values, timeout/cancel           |
| 5   | `lessons/05-static/`     | FileServer, 404/405, trailing slash      |
| 6   | `lessons/06-testing/`    | httptest without ports                   |

## Docs

- `docs/01-mux-vs-frameworks.md` — when stdlib wins, when gin/echo wins
