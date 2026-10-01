# Forwarding — the chain of who-told-whom

> TL;DR: proxies **append** (`X-Forwarded-For` grows right),
> origin trusts the chain only from **known hops**, reads the
> **leftmost** entry. Anything else is client graffiti.

## 1. Chain mechanics

```text
client (9.9.9.9) → proxy A → proxy B → origin
XFF: "9.9.9.9" → "9.9.9.9, A" → "9.9.9.9, A, B"
```

- Each honest proxy appends the **socket peer** it saw. Leftmost
  = original client, rightmost = last proxy before origin.
- `Forwarded: for=9.9.9.9;proto=https;by=B` is the RFC 7239
  successor (structured, same idea) — XFF won by incumbency.

## 2. Trust (the whole game)

- XFF is **client-writable**: case 2 shows a forged `1.2.3.4`
  arriving intact. Origin must decide per peer: known proxy →
  honor leftmost; anyone else → socket peer only.
- Attacker controls only their **own prefix**: through an honest
  proxy the forgery becomes `"1.2.3.4, real..."` — leftmost read
  still takes the lie. Full safety needs the edge proxy to
  **strip inbound XFF** (CDNs do this) so the chain starts clean.
- `X-Real-IP` (nginx): single value, overwritten per hop —
  simpler, but one proxy only; chains need XFF.

## 3. Rules

- Never trust XFF from the open internet (rate-limit, geo, audit
  on socket peer or a header your edge sets and strips).
- Edge strips or overwrites inbound XFF; inner hops append.
- Log both `remote_addr` and verdict `real_ip` — when spoof is
  suspected, the peer trail is the evidence.
