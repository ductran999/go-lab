.PHONY: default help build run up down migrate bench

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build binaries into ./bin (never bare go build)
	go build -o ./bin/demo ./cmd/demo
	go build -o ./bin/bench ./cmd/bench
	go build -o ./bin/migrate ./cmd/migrate

run: ## print v4 vs v7 side by side (no DB needed)
	go run ./cmd/demo

up: ## start postgres (:5438) for the pk bench
	docker compose up -d

down: ## stop postgres
	docker compose down

migrate: ## create t_int, t_u4, t_u7 (needs make up)
	go run ./cmd/migrate

bench: ## batch-insert 50k rows per PK shape (needs migrate)
	go run ./cmd/bench -n 50000
