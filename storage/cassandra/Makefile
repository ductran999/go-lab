.PHONY: default help build up down cql schema run

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build demo binary into ./bin (never bare go build)
	go build -o ./bin/demo ./cmd/demo

up: ## start cassandra (:9042, slow first boot ~1min)
	docker compose up -d

down: ## stop cassandra
	docker compose down

cql: ## open cqlsh
	docker compose exec cassandra cqlsh

schema: ## apply schema.cql (needs make up healthy)
	docker compose cp ./schema.cql cassandra:/schema.cql
	docker compose exec cassandra cqlsh -f /schema.cql

run: ## write + read demo (needs make schema first)
	go run ./cmd/demo
