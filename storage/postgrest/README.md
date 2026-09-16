# PostgREST Research

Stack: Postgres 16 + PostgREST + `migrate/migrate` (one-shot migration),
plus a small Go client demonstrating CRUD over REST.

```bash
# Change DIR
$ cd storage/postgrest
```

## 1. Setup

```bash
cp .env.example .env   # edit passwords if needed
make up                # docker compose up -d (db + migrator + postgrest)
make ps                # check containers
```

- Postgres: `localhost:5432` (`DB_NAME`/`DB_USER`/`DB_PASSWORD` from `.env`)
- PostgREST: `http://localhost:3000` (`PGRST_SCHEMA=api`, anon role `web_anon`)
- Migration: `migrations/000001_init.up.sql` creates schema `api`, table
  `api.todos`, roles `web_anon` + `authenticator`, and CRUD grants.
  A matching `.down.sql` file is provided for rollback.

> Note: the `authenticator` role password in the migration is hardcoded to
> `mysecretpassword` and must match `AUTHENTICATOR_PASSWORD` in `.env`
> (compose already uses this var for `PGRST_DB_URI`). To change the password,
> update both places.

## 2. Try the API with curl

```bash
make test-api
# GET
curl -s 'http://localhost:3000/todos?select=*'
# POST
curl -s -X POST 'http://localhost:3000/todos' \
  -H 'Content-Type: application/json' \
  -H 'Prefer: return=representation' \
  -d '{"task":"hello postgrest"}'
# PATCH
curl -s -X PATCH 'http://localhost:3000/todos?id=eq.1' \
  -H 'Content-Type: application/json' \
  -H 'Prefer: return=representation' \
  -d '{"done":true}'
# DELETE
curl -s -X DELETE 'http://localhost:3000/todos?id=eq.1'
```

## 3. Go client demo (clean architecture)

The client is split into layers with dependencies pointing inward:

```
cmd/main.go                          # delivery + composition root (wiring only)
internal/domain/todo.go              # entity + TodoRepository contract, no deps
internal/usecase/todo.go             # business rules (e.g. task validation)
internal/infrastructure/postgrest/   # PostgREST HTTP implementation (stdlib only)
```

`cmd/main.go` wires `postgrest.NewTodoRepository` into
`usecase.NewTodoUseCase` and runs one full CRUD cycle:
create → list → patch (mark done) → delete → list.
To target another data source, add a new `domain.TodoRepository`
implementation and swap the constructor in `cmd/main.go` — domain and
use case code stay untouched.

```bash
make run
# or
POSTGREST_URL=http://localhost:3000 go run ./cmd/main.go
```

Sample output:

```
== PostgREST CRUD demo ==
1. CREATED: {ID:2 Done:false Task:Learn PostgREST with Go Due:<nil>}
2. LIST (1 todos):
   - #2 done=false task="Learn PostgREST with Go"
3. UPDATED: {ID:2 Done:true Task:... Due:<nil>}
4. DELETED todo #2
5. REMAINING (0 todos)
```

## 4. Cleanup

```bash
make down    # stop, keep data
make clean   # stop + remove postgres_data volume
make logs    # tail postgrest logs
```
