.PHONY: default help build run-server run-client run-gobreaker

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build server, clients into ./bin (never bare go build)
	go build -o ./bin/server ./cmd/server
	go build -o ./bin/client ./cmd/client
	go build -o ./bin/gobreaker ./cmd/gobreaker

run-server: ## run killable downstream on :8124 (PORT to override)
	go run ./cmd/server

run-client: ## run dead-then-healed lifecycle (ARGS="-threshold 5 -cooldown 5s" to override)
	go run ./cmd/client $(ARGS)

run-gobreaker: ## same lifecycle on sony/gobreaker
	go run ./cmd/gobreaker $(ARGS)
