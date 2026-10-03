.PHONY: default help build run run-watchdog run-leader up down

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build binaries into ./bin (never bare go build)
	go build -o ./bin/cron ./cmd/cron
	go build -o ./bin/watchdog ./cmd/watchdog
	go build -o ./bin/leader ./cmd/leader

escape: ## show escape analysis (which vars move to heap, e.g. running)
	go build -gcflags=-m ./... 2>&1 | grep -E 'moved to heap|escapes' | head -10

run: ## run cron worker (@every 5s demo job, Ctrl-C to drain)
	go run ./cmd/cron

run-watchdog: ## run SLA watchdog sandbox (breach pages by design)
	go run ./cmd/watchdog

run-leader: ## run leader election demo (needs make up, run twice to see standby)
	PG_DSN='postgres://admin:password123@localhost:5439/jobsdb?sslmode=disable' go run ./cmd/leader

up: ## start postgres (:5439) for leader election
	docker compose up -d

down: ## stop postgres
	docker compose down
