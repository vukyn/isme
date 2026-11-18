# Makefile for isme

run:
	go run cmd/main.go

# Unit + database tests. The database tests start a throwaway Postgres
# container (kuery bun/pgtest), so Docker must be running; `-short` skips them.
test:
	go test ./...

migrate-up:
	go run db/migrate.go up

migrate-down:
	go run db/migrate.go down

migrate-reset:
	go run db/migrate.go reset

# Local Postgres (docker compose, host port 5433 — rainy uses 5432). Local flow:
# `make db-up`, then `make migrate-up` (a fresh database needs only that), then
# `make run`. The DB_* values in .env.example match the compose defaults.
.PHONY: test db-up db-down db-reset db-logs

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
	pnpm create vite@latest ui
	cd ui && pnpm add @chakra-ui/react @emotion/react
	cd ui && pnpm dlx @chakra-ui/cli snippet add

web:
	cd ui && pnpm run dev

build-web:
	cd ui && pnpm install && pnpm run build
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

# Repo-local process runner (mprocs.yaml in this directory).
.PHONY: mprocs
mprocs:
	mprocs
