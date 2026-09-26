# Forwarding lab — who is the client, really?

Origin on `:8103`, one honest proxy on `:8104`. Origin trusts
X-Forwarded-For only from known proxies (localhost here).

## Run

```bash
make run-origin  # terminal 1 — :8103
make run-proxy   # terminal 2 — :8104 → :8103
```

## Try

```bash
# 1. Direct: no headers, socket peer is the truth
curl -s localhost:8103/whoami | jq

# 2. Spoofed: client forges XFF, origin ignores it (peer untrusted)
curl -s localhost:8103/whoami -H 'X-Forwarded-For: 1.2.3.4' | jq .real_ip
# 127.0.0.1, NOT 1.2.3.4

# 3. Via proxy: peer is known, leftmost XFF honored
curl -s localhost:8104/whoami | jq

# 4. Spoof THROUGH proxy: forgery pushed right, truth stays left
curl -s localhost:8104/whoami -H 'X-Forwarded-For: 1.2.3.4' | \
  jq '{x_forwarded_for, real_ip}'
# "1.2.3.4, 127.0.0.1": attacker controls only their own prefix,
# never overwrites the proxy-appended suffix

# 5. Block: verdict enforced (BLOCKED_IPS matches trusted real_ip)
BLOCKED_IPS=127.0.0.1,::1 make run-origin  # restart origin, terminal 1
curl -s -o /dev/null -w '%{http_code}\n' localhost:8103/whoami
# 403 — hit ORIGIN directly: verdict is your real IP (::1/127.0.0.1).
# Spoofing XFF can't dodge it (still 403), because the blocklist
# reads the verdict, never raw headers.
# NOTE: spoofing THROUGH the proxy (:8104) takes leftmost and CAN
# dodge a naive list — that hole is the lesson, fixed by stripping
# inbound XFF at the edge (see docs).
```

## Docs

- `docs/01-forwarding.md` — XFF chain, trust, spoof mechanics
