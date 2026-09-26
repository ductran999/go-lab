.PHONY: default help build run-server run-web

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build server binary into ./bin (never bare go build)
	go build -o ./bin/server ./cmd/server

run-server: ## run auth lab API on :8098 (PORT to override)
	go run ./cmd/server

run-web: ## serve demo page on :8097 (cross-origin vs :8098)
	python3 -m http.server 8097 -d web
