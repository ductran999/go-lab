.PHONY: default help gen build run-server run-client

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

gen: ## regenerate protobuf code (needs buf + protoc plugins)
	buf generate --template buf.gen.yaml proto

build: ## build server and client into ./bin (never bare go build)
	go build -o ./bin/server ./cmd/server
	go build -o ./bin/client ./cmd/client

run-server: ## run grpc lab on :8107 (PORT to override)
	go run ./cmd/server

run-client: ## run client (create|get|watch, default watch)
	go run ./cmd/client $(CMD)
