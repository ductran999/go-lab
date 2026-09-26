.PHONY: default help build run-server run-client

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build server and client into ./bin (never bare go build)
	go build -o ./bin/server ./cmd/server
	go build -o ./bin/client ./cmd/client

run-server: ## run http3 lab on :8108/udp (PORT to override, certs from grpc lab)
	go run ./cmd/server

run-client: ## run client (needs server up + grpc lab CA present)
	go run ./cmd/client
