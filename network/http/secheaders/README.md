# Secheaders lab — same body, different armor

API on `:8105`. `/bare` vs `/hardened`: identical body, headers
do all the protecting.

## Run

```bash
make run-server  # API on :8105
```

## Try

```bash
# Bare: body only, zero protection
curl -s -D- -o /dev/null localhost:8105/bare

# Hardened: full armor set
curl -s -D- -o /dev/null localhost:8105/hardened
# X-Content-Type-Options, Content-Security-Policy (frame-ancestors),
# Referrer-Policy. HSTS appears only over TLS (by browser design).

# Attack demo (needs a real browser — curl can't execute):
# both serve the SAME script body as text/plain.
curl -s localhost:8105/polyglot-bare | head -c 120; echo
curl -s -D- -o /dev/null localhost:8105/polyglot-hardened | grep -i nosniff
# Now open both URLs in a browser:
# /polyglot-bare     → popup fires (browser sniffed text → HTML → XSS)
# /polyglot-hardened → inert text, no popup (nosniff held the line)
# NOTE (modern Chrome): top-level navigation no longer sniffs, so
# both show inert text. The live nosniff kill happens on
# SUBRESOURCES — open /demo instead:

# /demo loads the same JS twice: bare executes, nosniff is blocked
curl -s localhost:8105/lib-hard.js -D- -o /dev/null | grep -iE 'content-type|nosniff'
# Open http://localhost:8105/demo in Chrome:
# "bare script ran: true" vs "nosniff script ran: false" + console
# error "Refused to execute script ... MIME type ('text/plain') ...
# strict MIME checking is enabled".
```

## Docs

- `docs/01-security-headers.md` — attack → header mapping
