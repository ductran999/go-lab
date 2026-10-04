---
name: lab-layout
description: Use when creating a new lab or building/running inside one (cmd/server, Makefile targets, bin vs tmp outputs, ports).
---

# Lab layout

Every lab is self-contained and runnable from its own folder:

```
<lab>/README.md      Run + Try (curl/proof commands with expected output)
<lab>/Makefile       default/help/build/run-* (bench, up/down where needed)
<lab>/cmd/<name>/   mains (server, client, demo)
<lab>/internal/     lab-owned packages
<lab>/docs/         numbered research notes (see doc-table skill)
<lab>/bin/          release binaries (gitignored, never committed)
<lab>/tmp/          trial runs, scratch logs/pids (gitignored)
```

1. `go build -o ./bin/<name> ./cmd/<name>` — never bare `go build`.
   Trial/scratch binaries from agent experiments go to
   `<lab>/tmp/`, never beside the code, never repo-root `/tmp`.
2. Makefile: `.PHONY` + `default/help` stub, every target has a
   `## ` blurb (help greps it). Targets: `build`, `run-server`
   (PORT overridable), `run-client`/`bench` where they exist.
3. Server mains: `environ.Get("PORT", "<unique>")`, `fail(err)`
   helper, `slog`, `ReadHeaderTimeout: 5s`, atomic counters for
   anything the demo must prove with numbers.
4. Ports are unique per lab (`github.com/ductran999/shared-pkg/environ`
   default in code). Check before picking: grep `PORT", "81` repo-wide.
5. README Try section: every command shows expected output in a
   `#` comment line — the reader verifies without thinking.
6. Lab dependencies get their own `<lab>/docker-compose.yaml`
   (never a shared root compose): pinned image versions, ports
   that don't clash with other labs, `profiles:` for variants
   (direct/badger/persistent), config mounted `./x.yaml:/etc/x.yaml:ro`.
   Makefile `up`/`down` targets wrap it; README Run starts with
   `make up` when the lab needs it.
