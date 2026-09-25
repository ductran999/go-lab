# H2C REST lab — same dialect, faster vehicle

Plain REST JSON (`GET/POST /todos`) over cleartext HTTP/2 (`:8109`).
No protobuf, no stubs — the proof that H2 is transport, REST is dialect.

```bash
# Change DIR
$ cd network/http2
```

## Run

```bash
make run-server  # terminal 1 — :8109 h2c
make run-client  # terminal 2 — proto=HTTP/2.0 both calls
```

## TLS twin (same handlers, ALPN H2)

```bash
TLS_CERT=../grpc/certs/server.crt TLS_KEY=../grpc/certs/server.key make run-server &
TLS_CA=../grpc/certs/ca.crt make run-client
# proto=HTTP/2.0 over TLS now — curl with a modern build works too:
# curl --http2 https://localhost:8109/todos --cacert ../grpc/certs/ca.crt
```

`curl --http2-prior-knowledge` also speaks h2c (needs curl 7.47+;
ours lacks the flag — the Go client is the sure path).

## Docs

- `docs/01-rest-on-h2.md` — what changes (nothing) and what doesn't
