# Security — proving identity, hashing data

One pillar for safety: who you are (`auth/`) and how data is
sealed (`hashing/`).

## Map

| Dir        | What                                                                                               |
| ---------- | -------------------------------------------------------------------------------------------------- |
| `auth/`    | Identity flows: m2m, mTLS, OAuth2/OIDC, Goth, Keycloak                                             |
| `hashing/` | Hash benchmarks: passwords (argon2/bcrypt/scrypt), crypto (sha/blake3), non-crypto (xxhash/murmur) |

Ties: `../network/auth/` (token channels), `../network/secheaders/`
(response armor), `../practices/` (contract-first for APIs).
