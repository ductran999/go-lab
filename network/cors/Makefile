.PHONY: default help build run-server run-gin-demo run-web cache-up cache-down

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

run-server: ## run CORS lab API on :8090 (PORT to override)
	go run ./cmd/server

run-gin-demo: ## run same API with gin-contrib/cors on :8092 (PORT to override)
	go run ./cmd/gin-demo

build: ## build server binary into ./bin (never bare go build)
	go build -o ./bin/server ./cmd/server

run-web: ## serve demo page on :8091 (cross-origin vs :8090)
	python3 -m http.server 8091 -d web

cache-up: ## nginx cache on :8085 in front of :8090 (needs run-server)
	docker compose -f cache/docker-compose.yml up -d

cache-down: ## stop nginx cache
	docker compose -f cache/docker-compose.yml down
