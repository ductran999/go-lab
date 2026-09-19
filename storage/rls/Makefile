# Load .env if exists
ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: default help up down clean psql logs run-scoped run-rls run run-server token

TENANT ?= 1

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

up: ## start postgres + run migrations
	docker compose up -d

down: ## stop stack (keep volumes)
	docker compose down

clean: ## stop stack and remove volumes
	docker compose down -v

psql: ## open psql shell (admin, RLS bypassed: SET ROLE lab_user to test)
	docker exec -it pg-rls psql -U ${DB_USER} -d ${DB_NAME}

logs: ## tail postgres logs
	docker compose logs -f db

run-scoped: ## run manual-scoping demo (explicit WHERE per query)
	go run ./cmd/scoped

run-rls: ## run RLS demo (policy filtering, no WHERE)
	go run ./cmd/rls

run: ## run both demos back-to-back for comparison
	go run ./cmd/scoped && go run ./cmd/rls

run-server: ## serve comparison API on :8081 (see README for routes)
	go run ./cmd/server

token: ## print dev JWT for TENANT (default 1): make token TENANT=2
	@TENANT="$(TENANT)" python3 -c 'import base64,hashlib,hmac,json,os; s=os.environ["RLS_JWT_SECRET"]; t=int(os.environ.get("TENANT","1")); e=lambda d: base64.urlsafe_b64encode(json.dumps(d).encode()).rstrip(b"=").decode(); h=e({"alg":"HS256","typ":"JWT"}); p=e({"tenant_id":t}); print(h+"."+p+"."+base64.urlsafe_b64encode(hmac.new(s.encode(),(h+"."+p).encode(),hashlib.sha256).digest()).rstrip(b"=").decode())'
