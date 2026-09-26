# Frameworks — same API, four dialects

One Todos API implemented per framework. Compare routing, middleware,
binding, errors side by side — then pick by purpose, not religion.

| Dir | Stack | Port |
|-----|-------|------|
| `mux/` | stdlib only (deep-dive lessons inside) | :8112+ |
| `gin/` | gin-gonic/gin | :8120+ |
| `echo/` | labstack/echo | :8130+ |
| `fiber/` | gofiber/fiber (fasthttp) | :8140+ |

## Docs

- `mux/docs/01-mux-vs-frameworks.md` — control vs velocity
