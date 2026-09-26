.PHONY: default help build run-a run-b up down

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build both services into ./bin (never bare go build)
	go build -o ./bin/svc-a ./cmd/svc-a
	go build -o ./bin/svc-b ./cmd/svc-b

run-a: ## run entry service on :8110 (needs collector up)
	go run ./cmd/svc-a

run-b: ## run downstream service on :8111 (needs collector up)
	go run ./cmd/svc-b

up: ## start otel collector + jaeger (UI :16686)
	docker compose up -d

down: ## stop collector + jaeger
	docker compose down
