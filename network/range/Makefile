.PHONY: default help build run-server

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build server binary into ./bin (never bare go build)
	go build -o ./bin/server ./cmd/server

run-server: ## run range lab API on :8106 (PORT to override)
	go run ./cmd/server

fetch: ## parallel range download merged into ./tmp/lines.txt (server must run)
	go run ./cmd/fetch -parts 5 -out ./tmp/lines.txt && head -1 ./tmp/lines.txt && tail -1 ./tmp/lines.txt
