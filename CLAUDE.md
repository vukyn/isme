# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## The memory layer

@MEMORY.md

⚠️ **That import is the point of the file, not decoration.** `MEMORY.md` and
`memory/` are the distilled layer — one hard-won fact per file, with why it
matters — and they live **in the repository** because a machine's own Claude
memory directory is workspace-scoped and machine-local: this repo opened on
another machine, or outside the workspace the notes were written in, arrived with
none of them.

It is a **distillation, not the record.** This file and the repository's other
documents stay the authority; where a note disagrees with the file that owns the
subject, the repository wins and the note is what to fix. `MEMORY.md` carries the
rules the notes are written under — one line per note in the index, one fact per
file, say why rather than only what, and delete a wrong note rather than adding a
second one beside it.

## Common Commands

```bash
make db-up                  # local Postgres via docker compose (host port 5433)
make migrate-up             # go run db/migrate.go up (a fresh database needs only this)
make migrate-down           # rollback the last migration
make migrate-reset          # rollback every migration
make run                    # go run cmd/main.go
make build                  # build to ./bin
make test                   # go test ./... — database tests need Docker; -short skips them
make gen-key-rsa256         # generate certs/private.pem + public.pem (RSA 2048)

# UI (Vite + React + Chakra)
make web                    # cd ui && pnpm run dev
make build-web              # builds ui/dist into internal/ui (embedded by Go)

# Release
make tag VERSION=x.y.z      # creates + pushes git tag
```

Local flow: copy `.env.example` to `.env` (its DB_* values match the compose defaults),
then `make db-up` → `make migrate-up` → `make run`.

## Architecture

Clean architecture, domain-driven layout, Fiber HTTP, Bun ORM (Postgres), `sarulabs/di/v2` container.

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
  ⚠️ **A handler must NOT call `ctn.Delete()`** — it only *borrows* the container;
  `middlewares.DiContainerMiddleware` created it and releases it (see Dependency
  Injection below). This is the reverse of the old rule, which had every handler
  carry a `defer ctn.Delete()` (55 of them).
- Usecase never imports repository implementation — only the interface.
- Entities live in `entity/entity.go` with Bun tags + lifecycle hooks (`BeforeAppendModel`).

Existing domains: `user`, `auth`, `user_session`, `app_service`.

### Dependency Injection (`internal/di/`)

`di.NewBuilder()` aggregates definitions: `defineConfig`, `defineDB`, `defineMiddleware`, `defineRepository`, `defineUsecase`. Each returns `di.Def` with `Name`, `Build`. DI key constants live alongside (`di_repo`, `di_usecase`, `di_db`, `di_middleware`, `di_config`, `di_cache`).

Container is request-scoped via `middlewares.DiContainerMiddleware` in
`internal/middlewares/di_container_middleware.go`, mounted globally in
`internal/server/server.go` ahead of the routes.

⚠️ **`DiContainerMiddleware` owns the request container's whole lifetime.** It
creates the `di.Request` sub-container, stores it in Fiber locals, and releases it
with its own `defer` — so the release runs on every path: a handled 200, a 401 from
`Middleware.AuthMiddleware`, a 403 from `rbac.RequirePermission`, `/assets`,
`/favicon.svg`, a 404, every SPA catch-all render, a handler that returns an error,
and a panic recovered by `pkgRecover.NewFiberRecover()` (the recover middleware is
mounted INSIDE this one, and a `defer` runs during unwinding anyway).

**Whoever creates a sub-container releases it. Handlers only borrow one, so they must
not call `Delete`.** The rule used to be the opposite (a `defer ctn.Delete()` in each
of 55 handlers) and it leaked in production: this middleware is mounted globally,
ahead of the routes, so it had already built a sub-container by the time anything
decided the request would not reach a handler — and `sarulabs/di` keeps every
sub-container in its parent's `children` map until deleted, so each of those was
retained for the life of the **process**. A handler-owned lifetime cannot cover a
request that never reaches a handler. The leaking paths were also the cheapest,
unauthenticated ones (a tokenless 401, a 403, `/assets`, `/favicon.svg`, a 404, every
SPA render), so the leak was free to trigger. Measured in gardener, which carried the
byte-identical bug: live heap climbed monotonically to 58 MB over 60k unauthenticated
401s, versus 3–4 MB with the release in place.

It calls **`DeleteWithSubContainers`**, not `Delete`: `Delete` is conditional
(`containerSlayer.go:22-34`) — with any child present it merely sets
`deleteIfNoChild` and returns nil, leaving the container in the parent's `children`
map, i.e. the leak. A single owner needs an unconditional release. Its documented
hazard (tearing down a sub-container another goroutine still uses) does not apply:
nothing here touches the container after its handler returns — no
`SendStream`/`SetBodyStreamWriter`/`StreamRequestBody`, and no goroutine that resolves
from a request container (the only `go` statement outside tests is the Fiber listener
in `server.go`, and the scheduler deliberately builds its repositories straight off
the App-scoped DB). That is a **precondition of this design**: if you ever hand a
request-scoped dependency to a goroutine that outlives the request, this release is a
use-after-free and the ownership has to be rethought, not worked around.

⚠️ **`internal/di/di_middleware.go`'s sub-container is intentionally never deleted —
do not "fix" it.** It exists because the auth usecase is a `di.Request`-scoped
definition and a Request-scoped object cannot be resolved straight out of an `App`
container; the app-scoped `Middleware` then **retains** that usecase for the whole
process, so deleting the container it came from would run the usecase's `Close` and
destroy an object the middleware keeps calling. That is exactly ONE
permanently-retained container, created once at boot — bounded, unlike one per
request. It is also why `cmd/main.go` shuts down with `DeleteWithSubContainers()`: a
plain `Delete()` on the app container would be a no-op while that child exists.

`TestDiContainerMiddlewareReleasesRequestContainer`
(`internal/middlewares/di_container_middleware_test.go`) pins all of it: every path
above, `Close` called **exactly once** per request (so a re-added handler defer fails
the suite instead of silently double-closing — di's second `Delete` returns nil and
merely re-runs every registered `Close`, and isme's `Close` funcs are all debug log
lines, so nothing else would notice), and — the leak itself — that the app container
retains **zero** children afterwards. That last assertion needs no reflection:
`Delete()` on the parent only closes it when its `children` map is empty, so "the app
container closed on its first `Delete`" *is* "nothing was retained".

### Configuration

`internal/config/` uses `envconfig` + `godotenv`. `.env` at repo root drives:
- `APP_NAME`, `APP_ENV`, `APP_PORT`
- `AUTH_ACCESS_TOKEN_PRIVATE_KEY` + `AUTH_ACCESS_TOKEN_PUBLIC_KEY` (RS256, access token), `AUTH_REFRESH_TOKEN_SECRET_KEY` (HS256, refresh token), `AUTH_ACCESS_TOKEN_EXPIRE_IN`, `AUTH_REFRESH_TOKEN_EXPIRE_IN`
- `LOGGER_MODE`, `LOGGER_LEVEL`
- `GRACEFUL_*` (verbose, step delay, server shutdown timeout)
- `DB_*` (Postgres connection — see Database below)

`.env.example` lists every key with local defaults.

JWT: **access token = RS256** (signed with `AUTH_ACCESS_TOKEN_PRIVATE_KEY`, verified with `AUTH_ACCESS_TOKEN_PUBLIC_KEY`; consumers like medioa2/rainy verify with the public key only). **Refresh token = HS256** (`AUTH_REFRESH_TOKEN_SECRET_KEY`, isme-internal). Generate an RS256 keypair with `make gen-key-rsa256` (→ `certs/`).

### Database

**Postgres only** (bun `pgdialect` + `pgdriver`, via `kuery/bun/db`). Production runs on
Neon; locally `docker-compose.yml` starts `postgres:16` on host port **5433** (volume
`isme-pgdata`). Connection from `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME/DB_SSLMODE`, or
`DB_DSN`, which wins when set. Soft-delete via `deleted_at`; standard `created_at`/`updated_at`.

Migrations: `go run db/migrate.go up|down|reset` (it prints the target host and database
name, never the DSN). The history in `db/history/` is, in order:

1. **`000_baseline`** — the whole schema plus seed rows. ⚠️ **Frozen: never edit it.**
   Production recorded it as applied and will never re-run it, so an edit only reaches
   fresh installs and makes them drift from production.
2. **`001_…` – `034_…`** — no-op stubs (`legacy.go`) that keep the names of the old
   incremental history. Their effects all live in the baseline.
3. **The tail (`035_…` onwards)** — every schema or seed change is a new migration
   appended to `Migrations`, one per file.

⚠️ **Names are frozen.** The runner (`kuery/bun/migrate`) matches by name; a rollback fails
with "migration <name> not found" when an applied name leaves the slice. Never rename,
reorder or remove an entry — `TestMigrationNamesFrozen` pins the list. A database must
already hold `000_baseline` + `001…034` before `up` runs the tail (production does: ids
1-35); otherwise the stubs would be recorded without anything having run.

The SQLite era left `db/app.db` and `db/backups/` behind in old working copies. Nothing uses
them, but they hold real secrets: they stay gitignored — never commit or delete them.

**Tests** get a real database from `internal/testdb` (test-only), which wraps kuery
`bun/pgtest`: one throwaway Postgres container per test binary, the full history migrated
once into a template, and a fresh clone per `testdb.Open(t)`. Every package that opens a
database needs `func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }` (see any
`main_test.go`). Docker must be running — a missing daemon fails the tests rather than
skipping them — and `go test -short ./...` skips every database test. `db/history` tests
the migrations themselves on empty databases (`pgtest.EmptyDB`).

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
- `db/migrate.go` — migration CLI; `db/history/` — migrations
