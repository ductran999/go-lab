.PHONY: default help build run-01 run-02 run-03 run-04 run-05

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build lesson binaries into ./bin (never bare go build)
	go build -o ./bin/lesson01 ./lessons/01-routing
	go build -o ./bin/lesson02 ./lessons/02-middleware
	go build -o ./bin/lesson03 ./lessons/03-lifecycle
	go build -o ./bin/lesson04 ./lessons/04-context
	go build -o ./bin/lesson05 ./lessons/05-static

run-01: ## run lesson 01 routing on :8112
	go run ./lessons/01-routing

run-02: ## run lesson 02 middleware on :8113
	go run ./lessons/02-middleware

run-03: ## run lesson 03 lifecycle on :8114 (trap SIGINT for graceful shutdown)
	trap 'kill -INT $$' INT; go run ./lessons/03-lifecycle

run-04: ## run lesson 04 context on :8115
	go run ./lessons/04-context

run-05: ## run lesson 05 static on :8116
	go run ./lessons/05-static
