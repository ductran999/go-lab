# Auth — identity playground (fourth pillar)

How services and users prove who they are. Separate Go module
(`go-lab/auth`): own deps, root CI untouched. Complements
`../network/auth/` (token channels) — this folder is flows and
handshakes, that one is headers.

## Map

| Dir                 | What                                                        | Needs                                             |
| ------------------- | ----------------------------------------------------------- | ------------------------------------------------- |
| `m2m/`              | Machine-to-machine: auth service + billing, JWT propagation | —                                                 |
| `mtls/`             | Mutual TLS: server requires client certs, CN echoed back    | Local `*.key` files (gitignored, never committed) |
| `try-oauth-custom/` | Hand-rolled OAuth2 authorize + callback                     | Provider creds via env                            |
| `try-oidc-custom/`  | Hand-rolled OIDC on top of OAuth2                           | Provider creds via env                            |
| `try-goth/`         | Goth (Google) + sessions + JWT                              | `JWT_SESSION_KEY`, `JWT_SECRET_KEY`               |
| `play-keyloack/`    | Keycloak realm login flow                                   | `GO_BACKEND_CLIENT_ID/SECRET`                     |

## Run

```bash
cd auth
go mod tidy
go run ./mtls/server  # needs mtls/server/server.key next to it
```

Keys: demo-only, local-only. Regenerate with openssl if missing;
never commit `*.key` (gitleaks + `.gitignore` both watch).
