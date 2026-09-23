.PHONY: default help build run-server run-client

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build both binaries into ./bin (never bare go build)
	go build -o ./bin/server ./server
	go build -o ./bin/client ./client

run-server: ## run mTLS server on :8443 (needs server.key next to main.go)
	go run ./server

run-client: ## run mTLS client against :8443 (needs client.key next to main.go)
	go run ./client
