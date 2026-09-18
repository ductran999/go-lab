# Load .env if exists
ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: default help up down clean psql logs

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
