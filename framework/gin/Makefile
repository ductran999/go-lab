.PHONY: default help build run-server

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build binary into ./bin (never bare go build)
	go build -o ./bin/server ./cmd/server

run-server: ## run gin twin on :8120 (PORT to override)
	go run ./cmd/server
