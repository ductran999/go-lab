.PHONY: default tidy vet lint lint-fix build run-postgrest help

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

run-postgrest: ## run PostgREST CRUD demo (needs stack up, see storage/postgrest)
	$(MAKE) -C storage/postgrest run

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'
