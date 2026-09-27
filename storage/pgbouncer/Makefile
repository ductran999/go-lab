.PHONY: default help build up down init bench appbench conns conf psql

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## build demo binaries into ./bin (never bare go build)
	go build -o ./bin/demo ./cmd/demo
	go build -o ./bin/setleak ./cmd/setleak

up: ## start postgres (:5435) + pgbouncer (:6432)
	docker compose up -d

down: ## stop all
	docker compose down

init: ## create pgbench tables (once per fresh volume)
	pgbench -h localhost -p 5435 -U admin -d pooldb -i -s 10

bench: ## open N conns against direct vs pooled (needs pgbench + init)
	@echo "== direct :5435 =="
	pgbench -h localhost -p 5435 -U admin -d pooldb -c 50 -j 4 -T 10 -S
	@echo "== pooled :6432 =="
	pgbench -h localhost -p 6432 -U admin -d pooldb -c 50 -j 4 -T 10 -S

appbench: ## go app clients (50x20) direct vs pooled
	@echo "== direct :5435 =="
	go run ./cmd/demo -dsn 'postgres://admin:password123@localhost:5435/pooldb?sslmode=disable'
	@echo "== pooled :6432 =="
	go run ./cmd/demo -dsn 'postgres://admin:password123@localhost:6432/pooldb?sslmode=disable'

conns: ## show max_connections + live counts (direct vs pool)
	@echo "== direct :5435 =="
	PGPASSWORD=password123 psql -h localhost -p 5435 -U admin -d pooldb -tAc "SHOW max_connections;"
	PGPASSWORD=password123 psql -h localhost -p 5435 -U admin -d pooldb -c "SELECT state, count(*) FROM pg_stat_activity GROUP BY state;"
	@echo "== pool :6432 =="
	PGPASSWORD=password123 psql -h localhost -p 6432 -U admin -d pooldb -tAc "SHOW max_connections;" || true
	PGPASSWORD=password123 psql -h localhost -p 6432 pgbouncer -U admin -c "SHOW POOLS;" 2>/dev/null || echo "(admin console needs stats_users — see README)"

psql: ## open psql straight into postgres (:5435)
	PGPASSWORD=password123 psql -h localhost -p 5435 -U admin -d pooldb

setleak: ## prove SET leaks but SET LOCAL does not (needs pool up)
	go run ./cmd/setleak -dsn 'postgres://admin:password123@localhost:6432/pooldb?sslmode=disable'

conf: ## print effective values + the mounted postgresql.conf
	PGPASSWORD=password123 psql -h localhost -p 5435 -U admin -d pooldb -tAc \
		"SHOW max_connections; SHOW superuser_reserved_connections;" | \
		awk '{print "direct :5435 " $$0}'
	@echo "--- ./postgresql.conf (mounted) ---"
	@cat ./postgresql.conf
