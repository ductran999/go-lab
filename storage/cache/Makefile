.PHONY: default help build run bench bench-short up down

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build demo binary into ./bin (never bare go build)
	go build -o ./bin/demo .

run: ## run all backend examples (needs make up first)
	go run .

bench: ## full benchmark matrix (needs make up first, slow)
	go test -bench=. -benchmem .

bench-short: ## quick benchmark smoke (needs make up first)
	go test -bench=. -benchmem -benchtime=10x .

up: ## start redis, memcached, dragonfly, keydb
	docker compose up -d

down: ## stop backends
	docker compose down
