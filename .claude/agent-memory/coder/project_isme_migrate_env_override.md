---
name: isme-migrate-env-override
description: How to run isme migrate/seed against a scratch DB safely — env vars beat .env (godotenv no-override); .env DB_DSN may be prod; seed writes prod creds when APP_ENV=production
metadata:
  type: project
---

`config.LoadConfig` uses `godotenv.Load`, which never overrides a variable already set in the
process env. So `DB_SQLITE_PATH=<scratch>/x.db go run db/migrate.go sqlite up` and
`DB_DSN=postgres://...@127.0.0.1:<port>/... go run db/migrate.go postgres ...` hit only the
scratch target even though `.env` sets `DB_SQLITE_PATH`, `DB_DRIVER` and `DB_DSN`.

**Why:** `.env`'s `DB_DSN` may point at the real Postgres, and `cmd/seed` reads `APP_ENV` from
`.env` — with `production` it writes `../SEED_CREDENTIALS_PROD.md` in the platform root.
Verified 2026-11-18 by md5 of `db/app.db` and of the creds file before/after.

**How to apply:** always pass `DB_SQLITE_PATH` / `DB_DSN` (and `APP_ENV=development` for seed)
explicitly on the command line; prove the dev DB is untouched with an md5 before/after. A
disposable Postgres works via `docker run --rm -p 127.0.0.1:55432:5432 postgres:16-alpine`.
As of 2026-11-18 `cmd/seed` on Postgres fails at roles (`is_system` integer into BOOLEAN) —
a pre-existing bug, separate from the users upsert.
