.PHONY: default help build run-server run-client

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build server and client into ./bin (never bare go build)
	go build -o ./bin/server ./cmd/server
	go build -o ./bin/client ./cmd/client

run-server: ## run flaky downstream on :8121 (PORT to override)
	go run ./cmd/server

run-client: ## run the three-call demo (BASE to override)
	go run ./cmd/client
