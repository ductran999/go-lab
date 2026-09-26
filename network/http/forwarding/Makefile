.PHONY: default help build run-origin run-proxy

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build both binaries into ./bin (never bare go build)
	go build -o ./bin/origin ./cmd/origin
	go build -o ./bin/proxy ./cmd/proxy

run-origin: ## run origin on :8103 (PORT to override)
	go run ./cmd/origin

run-proxy: ## run proxy on :8104 forwarding to :8103 (ORIGIN to override)
	go run ./cmd/proxy
