# gRPC lab — typed contracts, binary wire, real streams

Todos service on `:8107` (HTTP/2 + protobuf). Unary Create/Get,
server-streaming Watch with `after_id` replay. Reflection on for
`grpcurl` (no protos needed to explore).

## Run

```bash
make run-server  # terminal 1 — :8107
make run-client CMD=create  # terminal 2
make run-client CMD=watch   # terminal 3 — streams until Ctrl-C
make run-client CMD=chat    # bidi echo
make run-client CMD=dual    # watch + chat on ONE connection (Wireshark lo!)
```

## Try

```bash
# grpcurl (install once): list, describe, call, stream
grpcurl -plaintext localhost:8107 list
grpcurl -plaintext localhost:8107 describe todo.v1.Todos
grpcurl -plaintext -d '{"task":"write lab"}' localhost:8107 todo.v1.Todos/Create
grpcurl -plaintext -d '{"after_id":0}' localhost:8107 todo.v1.Todos/Watch
```

## TLS posture (env switch)

```bash
# Plaintext (local demo): nothing set, insecure declared in code.
# TLS: set certs (7-day demo certs live in certs/, keys gitignored).
TLS_CERT=certs/server.crt TLS_KEY=certs/server.key make run-server &
GRPC_TLS_CA=certs/ca.crt make run-client CMD=create
# grpcurl over TLS: grpcurl -cacert certs/ca.crt ...
```

Regenerate expired certs with openssl (CA + localhost server cert);
never commit `*.key`.

## Docs

- `docs/01-grpc-basics.md` — HTTP/2, protobuf, 4 call types, vs REST/SSE
- `docs/02-protobuf-types.md` — scalar map, enums, well-known types
- `docs/03-http2.md` — H1 vs H2 plumbing, multiplexing, notes
