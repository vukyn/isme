---
name: isme-migrate-env-override
description: How to run isme migrate/seed against a scratch DB safely — env vars beat .env (godotenv no-override); .env DB_DSN may be prod; seed writes prod creds when APP_ENV=production
metadata:
  type: project
---

`config.LoadConfig` uses `godotenv.Load`, which never overrides a variable already set in the
process env. So `DB_DSN=postgres://...@127.0.0.1:<port>/... go run db/migrate.go up` hits only
the scratch target even though `.env` sets `DB_DSN`. (isme is Postgres-only since 2026-11-18:
no `sqlite`/`postgres` argument and no `DB_SQLITE_PATH` any more.)

**Why:** `.env`'s `DB_DSN` may point at the real Postgres, and `cmd/seed` reads `APP_ENV` from
`.env` — with `production` it writes `../SEED_CREDENTIALS_PROD.md` in the platform root.
Verified 2026-11-18 by md5 of `db/app.db` and of the creds file before/after.

**How to apply:** always pass `DB_DSN` (and `APP_ENV=development` for seed) explicitly on the
command line. `db/migrate.go` prints `Target: host=… database=…` first — check it before
trusting the run. A disposable Postgres works via
`docker run --rm -p 127.0.0.1:55432:5432 postgres:16-alpine`. The old `cmd/seed` roles bug
(`is_system` integer into BOOLEAN) was fixed 2026-11-18 by passing the Go bool.
