# Load .env if exists
ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: default tidy help up down clean logs ps run test-api token

TENANT ?= 1

default: ## show all available tasks
	@make help

tidy: ## install pkg listed in go.mod (run from repo root)
	cd ../.. && go mod tidy

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

up: ## start postgres + migrator + postgrest
	docker compose up -d

down: ## stop stack (keep volumes)
	docker compose down

clean: ## stop stack and remove volumes
	docker compose down -v

logs: ## tail postgrest logs
	docker compose logs -f postgrest

ps: ## list stack containers
	docker compose ps

run: ## run Go CRUD demo client (needs stack up)
	go run ./cmd/main.go

test-api: ## smoke test PostgREST with curl (GET + POST + DELETE)
	curl -s 'http://localhost:3000/todos?select=*' | head -c 500; echo
	curl -s -X POST 'http://localhost:3000/todos' \
		-H 'Content-Type: application/json' \
		-H 'Prefer: return=representation' \
		-d '{"task":"smoke test"}'; echo

token: ## print dev JWT for TENANT (default 1): make token TENANT=2
	@TENANT="$(TENANT)" python3 -c 'import base64,hashlib,hmac,json,os; s=os.environ["PGRST_JWT_SECRET"]; t=int(os.environ.get("TENANT","1")); e=lambda d: base64.urlsafe_b64encode(json.dumps(d).encode()).rstrip(b"=").decode(); h=e({"alg":"HS256","typ":"JWT"}); p=e({"role":"tenant_user","tenant_id":t}); print(h+"."+p+"."+base64.urlsafe_b64encode(hmac.new(s.encode(),(h+"."+p).encode(),hashlib.sha256).digest()).rstrip(b"=").decode())'
