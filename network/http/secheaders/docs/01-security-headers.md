# Security headers — cheap armor per response

> TL;DR: headers don't fix bugs, they **cap the blast radius**:
> **nosniff** kills type confusion, **CSP/frame-ancestors** kills
> framing, **HSTS** kills downgrade, **Referrer-Policy** caps leaks.

## 1. Attack → header

| Attack                                                          | Header                                                             | Effect                                        |
| --------------------------------------------------------------- | ------------------------------------------------------------------ | --------------------------------------------- |
| MIME sniffing (browser re-types `text/plain` → HTML, XSS fires) | `X-Content-Type-Options: nosniff`                                  | Declared type is final                        |
| Clickjacking (your page iframed under a fake button)            | `Content-Security-Policy: frame-ancestors 'none'` (or same-origin) | Refuses framing                               |
| Protocol downgrade (MITM strips HTTPS → HTTP)                   | `Strict-Transport-Security: max-age=31536000; includeSubDomains`   | Browser forces HTTPS for a year               |
| Referrer leak (`?token=` in URL sent to third party)            | `Referrer-Policy: strict-origin-when-cross-origin`                 | Cross-origin sends origin only, no path/query |

- `X-Frame-Options: DENY` is the legacy single-purpose version;
  CSP `frame-ancestors` supersedes it (use CSP, keep XFO only
  for ancient browsers).
- HSTS is honored **only over TLS** and sticks (browser caches
  the policy) — first visit stays exposed, hence preload lists.

## 2. Plain-language notes

- **`nosniff`**: server declares, browser obeys — no MIME
  guessing. Modern Chrome no longer sniffs top-level navigation
  anyway; the live kill is on **subresources** (`<script src>`
  with wrong MIME is refused — see `/demo`).
- **`Referrer-Policy: strict-origin-when-cross-origin`**:
  same-origin sends the full URL, cross-origin sends the bare
  origin — path/query (and `?token=`) never leak to third parties.
- **`HSTS`**: forces TLS. Browser upgrades every `http://` to
  `https://` for `max-age`, subdomains included — MITM downgrade
  to plaintext stops working.
- **CSP**: `default-src 'self'` locks resource loading to same
  origin (no outside code); `frame-ancestors 'none'` forbids
  anyone iframing the page (no clickjacking). Import gate +
  display ban in one header.

## 2. Rules

- Set at the **edge/gateway**, not per handler — one middleware
  for every response, no exceptions for error pages. Universally
  supported: nginx `add_header`, Envoy/Kong/ALB/Cloudflare
  toggles, framework middleware (helmet, gin). This lab sets
  them in code only to make them visible; production moves
  them to the edge.
- CSP is a program, not a flag: start `Report-Only`, collect
  violations, then enforce. Blind `default-src 'self'` breaks
  inline scripts and third-party widgets on day one.
- Headers cap damage; they never replace fixing XSS/mixed-content
  at the source.
