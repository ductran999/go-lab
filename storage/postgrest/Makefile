# Load .env if exists
ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: default tidy deps

default: ## show all available tasks
	@make help

tidy: ## install pkg listed in go.mod
	go mod tidy

help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

up:
	docker compose up -d

down:
	docker compose down 

clean:
	docker compose down -v

run:
	go run ./cmd/main.go
