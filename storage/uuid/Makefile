.PHONY: default help build run

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build demo binary into ./bin (never bare go build)
	go build -o ./bin/demo ./cmd/demo

run: ## print v4 vs v7 side by side
	go run ./cmd/demo
