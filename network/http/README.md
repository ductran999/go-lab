# HTTP labs — grouped by header cluster

```bash
# Change DIR
$ cd network/http
```

| Headers                                   | Lab            | Core lesson                     |
| ----------------------------------------- | -------------- | ------------------------------- |
| `Origin`, `Access-Control-*`              | `cors/`        | Preflight, credentials          |
| `Authorization`, `Cookie`, `Set-Cookie`   | `auth/`        | Bearer vs cookie vs query       |
| `Cache-Control`, `ETag`, `Vary`           | `cache/`       | Freshness vs validation         |
| `Accept`, `Content-Type`                  | `negotiation/` | q-ranking, 406                  |
| `X-Forwarded-For`, `Forwarded`            | `forwarding/`  | Trust by peer                   |
| `Content-Security-Policy`, `HSTS`, ...    | `secheaders/`  | Blast-radius caps               |
| `Range`, `Accept-Ranges`, `Content-Range` | `range/` | Slices + resume |
| — (mechanics) | `../../framework/` | mux/gin/echo/fiber side by side |
