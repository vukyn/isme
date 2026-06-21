# Makefile for rainy

run:
	go run cmd/main.go

migrate-up:
	go run db/migrate.go $(DB) up

migrate-down:
	go run db/migrate.go $(DB) down

migrate-reset:
	go run db/migrate.go $(DB) reset

# Apply the squashed dual-dialect baseline as the fresh-install path (full
# schema + seed + stamp 001-029). e.g. make migrate-baseline DB=postgres
migrate-baseline:
	go run db/migrate.go $(DB) baseline

# Local Postgres for DB_DRIVER=postgres (docker compose). Dev-only infra — isme
# itself still runs via `make run`. Host port 5433 (rainy uses 5432). After
# `make db-up`, uncomment the Postgres block in .env (DB_DRIVER=postgres ...),
# then `make migrate-up DB=postgres && make run`.
.PHONY: db-up db-down db-reset db-logs

db-up:
	docker compose up -d postgres
	@echo "Waiting for Postgres to be ready..."
	@until docker compose exec -T postgres pg_isready -U $${DB_USER:-isme} >/dev/null 2>&1; do \
		sleep 1; \
	done
	@echo "Postgres ready on 127.0.0.1:$${DB_PORT:-5433} (db=$${DB_NAME:-isme} user=$${DB_USER:-isme})"

db-down:
	docker compose down

# db-reset also removes the data volume — WIPES all local Postgres data.
db-reset:
	docker compose down -v

db-logs:
	docker compose logs -f postgres

gen-key-rsa256:
	mkdir -p certs
	openssl genpkey -algorithm RSA -out certs/private.pem -pkeyopt rsa_keygen_bits:2048
	openssl rsa -pubout -in certs/private.pem -out certs/public.pem

init-ui:
	npm create vite@latest ui
	cd ui && npm i @chakra-ui/react @emotion/react
	cd ui && npx @chakra-ui/cli snippet add

web:
	cd ui && npm run dev

build-web:
	cd ui && npm install && npm run build
	rm -rf ./internal/web/dist
	mv ./ui/dist ./internal/web/dist
	touch ./internal/web/dist/.gitkeep

v-tag:
	git tag -l --sort=-version:refname

v-tag-latest:
	git tag -l --sort=-version:refname | head -n 1

tag:
	git tag -a v$(VERSION) -m "Release version $(VERSION)"
	git push origin v$(VERSION)