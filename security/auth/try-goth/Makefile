.PHONY: default help build run

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build binary into ./bin (never bare go build)
	go build -o ./bin/demo .

run: ## run Goth (Google) demo (needs SERVER_HOST, JWT_SESSION_KEY, JWT_SECRET_KEY)
	go run .
