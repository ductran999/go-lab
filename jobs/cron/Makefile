.PHONY: default help build run

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build binary into ./bin (never bare go build)
	go build -o ./bin/cron ./cmd/cron

escape: ## show escape analysis (which vars move to heap, e.g. running)
	go build -gcflags=-m ./... 2>&1 | grep -E 'moved to heap|escapes' | head -10

run: ## run cron worker (@every 5s demo job, Ctrl-C to drain)
	go run ./cmd/cron
