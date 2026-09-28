.PHONY: default help up down load fetch report openrep clean

default: ## show all available tasks
	@make help

help: ## show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

up: ## start dedicated postgres (:5437) with analysis logging
	docker compose up -d

down: ## stop postgres (logs vanish with it — fetch first!)
	docker compose down

load: ## generate traffic (pgbench init + run on :5437)
	PGPASSWORD=password123 pgbench -h localhost -p 5437 -U admin -d badgerdb -i -s 5 2>/dev/null || true
	PGPASSWORD=password123 pgbench -h localhost -p 5437 -U admin -d badgerdb -c 10 -j 2 -T 15

fetch: ## pull db container logs into ./tmp/pg.log
	mkdir -p ./tmp
	docker compose logs --no-log-prefix db > ./tmp/pg.log 2>/dev/null
	wc -l ./tmp/pg.log

report: ## analyze ./tmp/pg.log into ./tmp/report/out.html (dalibo/pgbadger)
	mkdir -p ./tmp/report
	docker run --rm -v ./tmp:/tmp dalibo/pgbadger \
		-O /tmp/report/ -f stderr /tmp/pg.log
	@echo "open ./tmp/report/out.html"

openrep: ## open the report in the default browser
	xdg-open ./tmp/report/out.html

clean: ## drop fetched logs and reports
	rm -rf ./tmp
