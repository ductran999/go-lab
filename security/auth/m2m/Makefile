.PHONY: default help build run-auth run-billing

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build both binaries into ./bin (never bare go build)
	go build -o ./bin/auth ./auth
	go build -o ./bin/billing ./billing

run-auth: ## run auth service on :8080
	go run ./auth

run-billing: ## run billing service (calls auth on :8080)
	go run ./billing
