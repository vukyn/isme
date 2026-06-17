# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

```bash
make run                    # go run cmd/main.go
make build                  # build to ./bin
make migrate-up DB=sqlite   # run db/migrate.go <DB> up (dbType is "sqlite", not "app")
make migrate-down DB=sqlite # rollback
make migrate-reset DB=sqlite # reset
make gen-key-rsa256         # generate certs/private.pem + public.pem (RSA 2048)

# UI (Vite + React + Chakra)
make web                    # cd ui && npm run dev
make build-web              # builds ui/dist into internal/ui (embedded by Go)

# Release
make tag VERSION=x.y.z      # creates + pushes git tag
```

No `go test` targets defined. Run `go test ./...` directly when adding tests.

## Architecture

Clean architecture, domain-driven layout, Fiber HTTP, Bun ORM (SQLite), `sarulabs/di/v2` container.

Entry: `cmd/main.go` → `internal/app` → `internal/server` (Fiber) + DI build.

### Layer flow per domain (`internal/domains/<domain>/`)

```
handlers/http  →  usecase  →  repository  →  entity
                              external/*       (DB)
models/         request + response DTOs
constants/      domain constants
exceptions/     domain errors
```

Rules:
- Repository implements interface in `irepository.go` (`type IRepository interface`); same for `iusecase.go`.
- Handlers receive deps via DI container injected by middleware; resolve with `di.Get(ctx, key)`.
- Usecase never imports repository implementation — only the interface.
- Entities live in `entity/entity.go` with Bun tags + lifecycle hooks (`BeforeAppendModel`).

Existing domains: `user`, `auth`, `user_session`, `app_service`.

### Dependency Injection (`internal/di/`)

`di.NewBuilder()` aggregates definitions: `defineConfig`, `defineDB`, `defineMiddleware`, `defineRepository`, `defineUsecase`. Each returns `di.Def` with `Name`, `Build`. DI key constants live alongside (`di_repo`, `di_usecase`, `di_db`, `di_middleware`, `di_config`, `di_cache`).

Container is request-scoped via middleware in `internal/middlewares/`.

### Configuration

`internal/config/` uses `envconfig` + `godotenv`. `.env` at repo root drives:
- `APP_NAME`, `APP_ENV`, `APP_PORT`
- `AUTH_ACCESS_TOKEN_SECRET_KEY`, `AUTH_REFRESH_TOKEN_SECRET_KEY`, `AUTH_ACCESS_TOKEN_EXPIRE_IN`, `AUTH_REFRESH_TOKEN_EXPIRE_IN`
- `LOGGER_MODE`, `LOGGER_LEVEL`
- `GRACEFUL_*` (verbose, step delay, server shutdown timeout)

JWT signing uses HS256 with the two secret keys above (RSA certs in `certs/` reserved for future use via `gen-key-rsa256`).

### Database

SQLite at `db/app.db`. Migrations: `go run db/migrate.go <db-name> up|down|reset`. Migration history under `db/history/`. Bun dialect: `sqlitedialect` with `sqliteshim` driver. Soft-delete via `deleted_at`; standard `created_at`/`updated_at`.

### External services

`external/auth/services/` — outbound HTTP client (resty) for an external auth provider. Constants in `external/auth/constants/api.go`. Used by usecase layer, not repository.

### UI

Vite + React + Chakra UI in `ui/`. `make build-web` compiles and moves `ui/dist` → `internal/ui` for Go embedding (Fiber serves the SPA + assets). `internal/ui/assets` is build output — do not hand-edit.

### Shared packages (`github.com/vukyn/kuery`)

The old local `pkg/` directory was consolidated into the `github.com/vukyn/kuery` module (≥ v1.12.0):
- `kuery/jwt` — token gen/validate (HS256 + RS256)
- `kuery/claims` — JWT claim shape
- `kuery/ctx` — typed context keys + helpers (e.g. user_id from claims)
- `kuery/http/fiber` — standardized JSON responses + error mapping
- `kuery/graceful` — shutdown coordinator
- `kuery/recover` — panic recovery middleware
- `kuery/bun/{hooks,query}` — ORM helpers
- `kuery/cryp` — crypto helpers

## Conventions

- Package names: lowercase, no underscores. Files: `snake_case.go`.
- Interfaces prefixed `I` (`IUserRepository`); files `irepository.go`, `iusecase.go`.
- Constants: `UPPER_SNAKE_CASE`.
- Use early returns; keep functions small.
- Bun tags style: `bun:"id,pk,autoincrement"`. JSON tags on response models.
- Errors: domain exceptions in `exceptions/`, mapped by handler via `kuery/http/fiber`. Don't return raw `error` to clients.
- Logging: structured via zerolog (Fiber middleware `fiberzerolog`).
- Imports grouped: stdlib, third-party, internal.
- Frontend: `docs/frontend-structure.md` (ui/src layout) + `docs/chakra-v3.md` (Chakra v3 only, never v2 syntax).

## Key References

- `cmd/main.go` — bootstrap
- `internal/app/app.go` — app init
- `internal/server/server.go` — Fiber + routes
- `internal/di/di.go` — DI builder
- `internal/domains/user/` — canonical domain example
- `db/migrate.go` — migration runner
