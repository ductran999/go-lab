# HATEOAS — server guides, client follows

> TL;DR: Level 3 REST embeds **next-step links** in responses so
> clients follow **state, not hardcoded URLs**. Elegant, rarely
> worth it — Level 2 + OpenAPI won in practice.

## 1. Core idea

Instead of clients hardcoding flows from docs, the server attaches
`_links` describing valid next actions for the current state:

```json
// GET /api/orders/123 → pending_payment
{"order_id": 123, "status": "pending_payment",
 "_links": {"self": {"href": "/api/orders/123", "method": "GET"},
            "pay": {"href": "/api/orders/123/pay", "method": "POST"},
            "cancel": {"href": "/api/orders/123/cancel", "method": "POST"}}}

// same URL → delivered: links change with state
{"order_id": 123, "status": "delivered",
 "_links": {"self": {"href": "/api/orders/123", "method": "GET"},
            "refund": {"href": "/api/orders/123/refund", "method": "POST"}}}
```

## 2. Trade-offs

| Pros | Cons |
|------|------|
| Decoupled: URL changes don't break clients | Backend owns state machine + link generation (heavy) |
| Self-navigating responses | Payload bloat (links on every resource) |

## 3. Verdict

- Most teams stop at Level 2 + Swagger/OpenAPI, or move to
  GraphQL/gRPC. HATEOAS pays off only when many heterogeneous
  clients must survive server-driven flow changes.
- Ties to `01-contract-first-vs-code-first.md`: HATEOAS is
  contract-first taken to runtime — the contract travels inside
  responses instead of a spec file.
