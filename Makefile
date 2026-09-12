.PHONY: default tidy vet lint lint-fix build help

default: ## show all available tasks
	@make help

tidy: ## tidy go.mod dependencies
	go mod tidy

vet: ## run go vet on all packages
	go vet ./...

lint: ## run golangci-lint on all packages
	golangci-lint run ./...

lint-fix: ## run golangci-lint with auto-fix on all packages
	golangci-lint run --fix ./...

build: ## build all packages
	go build ./...

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'
