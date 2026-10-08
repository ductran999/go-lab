.PHONY: default help build run-server run-client up down

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build server and client into ./bin (never bare go build)
	go build -o ./bin/server ./cmd/server
	go build -o ./bin/client ./cmd/client

run-server: ## run killable primary on :8127 (PORT to override)
	go run ./cmd/server

run-client: ## walk the ladder (ARGS="-n 3 -strict" to override)
	go run ./cmd/client $(ARGS)

up: ## start shared redis on :6379
	docker compose up -d

down: ## stop shared redis
	docker compose down
