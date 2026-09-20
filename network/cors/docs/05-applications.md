# Real-World Applications

**TL;DR:** CORS is invisible when right, blocking when wrong. These patterns cover most production setups.

Keywords: session persistence, credentialed APIs, CDN.

## Persistent login (remember me)

- First visit: login → `Set-Cookie` (HttpOnly) → jar.
- Every later visit: app calls `GET /me` with `credentials: 'include'`
  on load. Valid session → straight to home, no login form.
  Expired/missing → 401 → stay on login.
- Requires: exact ACAO echo + `Allow-Credentials: true`. `*` breaks it.

## Credentialed cross-origin APIs

- SaaS frontend (`app.example.com`) + API (`api.example.com`):
  cookies flow via `include`, CORS headers gate JS reads.
- Same pattern powers multi-subdomain SSO: one session cookie,
  every service validates it independently.

## Public APIs

- No cookies, no auth headers → plain `*`, no credentials.
  Cheapest correct setup. Adding credentials later means switching
  to echo + `Vary` — plan the migration, don't bolt it on.

## CDN in front

- Cache key must include Origin (or backend sends `Vary: Origin`),
  else one tenant's ACAO poisons everyone (see `03-cache-vary.md`).
- Strip `Server` version + debug headers at the edge.
