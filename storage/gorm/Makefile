.PHONY: default help build run up down migrate

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build demo binaries into ./bin (never bare go build)
	go build -o ./bin/demo ./cmd/demo
	go build -o ./bin/migrate ./cmd/migrate

up: ## start direct postgres (:5436)
	docker compose up -d

down: ## stop postgres
	docker compose down

migrate: ## create users table (needs make up)
	go run ./cmd/migrate

run: ## print generated SQL per Session flag (needs make up, writes nothing)
	go run ./cmd/demo
