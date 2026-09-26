# Auth — identity playground (fourth pillar)

How services and users prove who they are. Same Go module as the
repo root: `make tidy` covers these deps too. Complements
`../../../network/http/auth/` (token channels) — this folder is flows and
handshakes, that one is headers.

## Map

| Dir                 | What                                                        | Run                                              |
| ------------------- | ----------------------------------------------------------- | ------------------------------------------------ |
| `m2m/`              | Machine-to-machine: auth service + billing, JWT propagation | `make -C m2m run-auth` + `run-billing`           |
| `mtls/`             | Mutual TLS: server requires client certs, CN echoed back    | `make -C mtls run-server` + `run-client`         |
| `try-oauth-custom/` | Hand-rolled OAuth2 authorize + callback (`:8080`)           | `make -C try-oauth-custom run`                   |
| `try-oidc-custom/`  | Hand-rolled OIDC on top of OAuth2 (`:8080`)                 | `make -C try-oidc-custom run`                    |
| `try-goth/`         | Goth (Google) + sessions + JWT                              | `make -C try-goth run` (`SERVER_HOST`, JWT keys) |
| `play-keyloack/`    | Keycloak realm login flow (`:8081`)                         | Needs Keycloak on `:8080` + client ID/secret     |

## Run

```bash
make -C auth/mtls run-server  # from repo root; needs server.key next to main.go
```

Keys: demo-only, local-only. Regenerate with openssl if missing;
never commit `*.key` (gitleaks + `.gitignore` both watch).
