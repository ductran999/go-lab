-include .env
export $(shell sed 's/=.*//' .env 2>/dev/null)

default: help

help: ## Show help for each of the Makefile commands
	@awk 'BEGIN \
		{FS = ":.*##"; printf "Usage: make ${cyan}<command>\n${white}Commands:\n"} \
		/^[a-zA-Z_-]+:.*?##/ \
		{ printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' \
		$(MAKEFILE_LIST)

.PHONY: tidy
tidy: ## Tidy up the go.mod
	go mod tidy

.PHONY: lint
lint: ## Run linters
	golangci-lint run --timeout 10m --config .golangci.yml

.PHONY: kafka
kafka: ## Start demo
	@docker compose up -d broker kafka-ui

rabbitmq: ## Start rabbitmq
	@docker compose up -d rabbitmq

.PHONY: cleanup
cleanup: ## Cleanup demo
	@docker compose down

APP_NAME := xplr-distributed-mq
MODULE   := xplr-distributed-mq

VERSION   ?= 1.0.0
GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build:
	go build \
		-ldflags "\
		-X $(MODULE)/cmd.Version=$(VERSION) \
		-X $(MODULE)/cmd.GitCommit=$(GIT_COMMIT) \
		-X $(MODULE)/cmd.BuildDate=$(BUILD_DATE)" \
		-o xplr-mq
