# Plan: R2 off-host backup + restore (isme)

Status: **PLANNED** (not started). Extends the shipped local `database_backup` job (PR #61).
Author: planning session 2026-06-17.

## Goal

Extend isme's existing `database_backup` scheduled job with:
- **(A)** after the local `VACUUM INTO` snapshot, optionally push the backup file **privately** off-host to Cloudflare R2, tracking the object key in `last_result` and pruning old R2 objects by count.
- **(B)** a **restore** capability that lists available backups (local + R2) and restores a chosen one safely.

Restore is delivered as an **operator CLI run with the service stopped**, not a one-click in-app endpoint, because the live app holds `db/app.db` open via the App-scoped `*bun.DB` and an in-process file swap risks corruption.

## What already exists (on main, PR #61)

Scheduled `database_backup` job:
- `internal/di/scheduler_jobs.go` `newDatabaseBackupRun` → SQLite `VACUUM INTO db/backups/app-<UTC-ts>.db` then `pruneBackups` (default retain 10).
- Config in generic `schedule_config` table (job_key `database_backup`, params `{retain_count}`, last_result JSON `{backup_path,kept,deleted,bytes}`).
- Settings UI: `ui/src/components/DatabaseBackupCard.tsx`.
- JobKey const: `internal/domains/settings/entity/schedule_config.go`.
- Backup primitive is isme-local (not kuery).

## ⚠️ Critical security finding: how isme pushes a private .db to R2

The kuery/medioa SDK (`kuery v1.27.0`, `medioa/upload.go`) exposes **only public upload**. The medioa2 read endpoint `GET /api/v1/public/objects/{token}` is **anonymous — gated only by an unguessable SHA-256 token, no `X-API-Key`** (by design, so browsers can load `<img>`/`<audio>`).

A DB backup contains password hashes and app secrets. Storing it behind an anonymously-readable token is a security regression. **The medioa public path is NOT acceptable for backups.**

### Options
- **(a) Existing kuery/medioa SDK** — REJECTED. No private-upload; read side world-readable-by-token.
- **(b) isme direct R2 + aws-sdk-go-v2** — **RECOMMENDED.** `PutObject` to a private bucket/prefix isme controls; never public; restore reads back with the same creds. Cost: new dep + new secrets in isme `.env`. Mirrors how medioa2 talks to R2. The R2 client wrapper goes into **kuery** (`kuery/r2`) as a versioned package (reusable by medioa2/rainy).
- **(c) New authenticated medioa2 service-to-service blob endpoint** — REJECTED for now. Highest coupling + surface; medioa2 would take custody of another service's secret-bearing backups. Defer unless medioa2 becomes the single R2 gateway for the platform.

**Recommendation: Option (b).** Backups must be private; isme owns its own secrets boundary; restore needs read-back; lowest coupling.

## Decisions needed (recommendation in bold)
1. **Confirm Option (b)** (isme direct R2) vs (c) new private medioa2 endpoint. *Blocks PR-A1.*
2. **R2 creds: a separate private backup bucket + its own key** (blast-radius isolation from medioa2's media bucket) vs reuse medioa2's bucket under `isme-backups/` prefix. *Blocks PR-A1.*
3. **R2 retention: in-job count prune** (self-contained) vs R2 bucket lifecycle policy.
4. **Restore: CLI-with-service-stopped** (safe) vs full in-app maintenance-mode swap (risky). *Blocks PR-B.*
5. **Seed migration: code-default `r2_enabled=false`, no migration** vs explicit `030_*` seed.

---

## Phase PR-kuery — `kuery/r2` client (ship first)

New kuery package (separate PR + tag, then `go get` in isme):
- `kuery/r2/client.go` — `New(Config{AccountID/Endpoint, AccessKeyID, SecretAccessKey, Bucket})` wrapping `aws-sdk-go-v2` S3 client for R2 (custom endpoint resolver, `region="auto"`). Methods: `Put(ctx, key, r io.Reader, size int64, contentType string) error`, `Get(ctx, key) (io.ReadCloser, error)`, `List(ctx, prefix) ([]Object{Key,Size,LastModified}, error)`, `Delete(ctx, key) error`. Generic (no "backup" semantics) for reuse.
- Commit → `make tag VERSION=1.28.0` → delete tags older than 5 newest (local+remote) → `go get github.com/vukyn/kuery/r2@v1.28.0` in isme.

## Phase PR-A1 — upload-to-R2 in the backup job

**isme config (`internal/config/config.go`, near the `Medioa` struct):** add an `R2` struct — `AccountID`/`Endpoint`, `AccessKeyID`, `SecretAccessKey`, `Bucket`, `Prefix` (default `isme-backups/`), via `envconfig`; blank by default so isme boots without R2.
`.env` (root, gitignored — never commit; prod = fly secrets): `R2_ACCOUNT_ID`/`R2_ENDPOINT`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BACKUP_BUCKET`, `R2_BACKUP_PREFIX`.

**isme DI (`internal/di/di_service.go` + `di.go`):** mirror `defineMedioaClient` — add `defineR2Client()` (App scope, typed-nil when creds blank, Warn log), `GetR2Client(ctn)`, constant `CONTAINER_NAME_R2_CLIENT` in `internal/constants/di.go`. Register in `di.go` aggregation.

**Backup job (`internal/di/scheduler_jobs.go`):**
- Extend `databaseBackupParams` with `R2Enabled bool` (and optional `R2RetainCount int64`, default to `RetainCount` if zero).
- `newDatabaseBackupRun` also receives the R2 client (resolved in the closure that builds the job, like `db`). Inject an **interface** (not the concrete client) for mockability.
- After local VACUUM + `pruneBackups`: if `R2Enabled` and client non-nil → open `target`, `r2.Put(ctx, prefix+filepath.Base(target), file, bytes, "application/x-sqlite3")`; record `r2_key`+`r2_uploaded=true`.
- **Failure semantics:** R2 error logged + recorded `r2_uploaded=false`+`r2_error`, job returns `nil`. R2 failure must NOT fail the already-successful local backup (matches never-panic convention).
- Extend `last_result` with `r2_uploaded bool`, `r2_key string`, `r2_error string` (omitempty); update the result struct read by the settings usecase.

**Verify:** `go test ./...` (+ job-body unit test with fake R2 interface); manual: real creds → enable r2 → trigger → private object lands + NOT anonymously fetchable + `last_result.r2_uploaded=true`; negative: bad creds → local backup still ok + `r2_error` set + job nil.

## Phase PR-A2 — R2 retention + Settings UI

**Backend:** in-job R2 prune — `r2.List(prefix)`, sort by key/LastModified, `r2.Delete` past `R2RetainCount`; record `r2_deleted`/`r2_kept`. Extend `DatabaseBackupUpdateRequest`/`GetResponse` (`settings/models/database_backup.go`) with `R2Enabled` (+ `R2RetainCount` if exposed); bound in `Validate()`; update usecase Get/Update. Surface `last_r2_uploaded`/`last_r2_key` in the response.

**Frontend (`DatabaseBackupCard.tsx` + `ui/src/apis`):** add "Off-host backup (R2)" toggle (mirror enable switch) wired to dirty/discard/save; add R2 status to the last-run strip; extend api types. Chakra v3 only. After: `make build-web` + restart Go server (embedded SPA).

**Verify:** `go test ./...`, `npm run lint`, `make build-web`; manual toggle round-trip + R2 prune keeps N newest; RBAC unchanged (`settings:read/update`).

## Phase PR-B — restore (operator CLI + read-only list)

### Mechanism: CLI with service STOPPED. NOT in-app one-click.

**Why in-app one-click is unsafe:** isme holds `db/app.db` open through the App-scoped `*bun.DB` singleton shared by every request sub-container; SQLite keeps the file (+ `-wal`/`-shm`) open for the process lifetime. Overwriting under live connections corrupts in-flight txns + WAL; open handles point at the replaced inode. In-process would need maintenance-mode + request drain + `db.Close()` + checkpoint(TRUNCATE) + delete wal/shm + swap + re-open + re-register across the App container — fragile. Safe path = CLI with service stopped + UI that only **lists** and shows the copy-paste command.

**New CLI `cmd/restore/main.go`** (sibling to `cmd/main.go`, migrate.go style — plain main, no DI/Fiber):
- `go run cmd/restore/main.go list` — enumerate local `db/backups/app-*.db` (reuse `pruneBackups` glob/sort) + (if R2 creds present) `r2.List(prefix)`. Print key, size, mtime.
- `go run cmd/restore/main.go restore <source>` where `<source>` = local path OR `r2:<key>`:
  1. R2 source → `r2.Get` → temp path.
  2. **Validate:** open read-only, `PRAGMA integrity_check` == `ok`; assert core tables exist (`schedule_config`/`users`). Abort on failure — never overwrite from a bad file.
  3. **Snapshot current first** (reversibility): copy `db/app.db` → `db/backups/pre-restore-<ts>.db`; best-effort abort if server appears running.
  4. Atomic swap: write `db/app.db.tmp` → `os.Rename`; remove stale `db/app.db-wal`/`-shm`.
  5. Print "restore complete — start the service".

**UI (read-only list, NO restore button):** new `GET /settings/database-backup/restore-points` (handler + route, `PERM_SETTINGS_READ`) listing local + R2 backups; usecase `ListRestorePoints` (filesystem + R2 client, no repo). Render list + for selected item show exact command `go run cmd/restore/main.go restore <source>`.

**Auth:** list endpoint = `settings:read` (only filenames/sizes). Destructive CLI = implicitly operator/superadmin (needs host shell + R2 creds; isme superadmin = admin-on-isme-app). If a destructive endpoint is ever exposed, gate superadmin-only, stronger than `settings:update` — default plan exposes none.

**Verify:** `go test ./...` (+ integrity-check good/corrupt + safe-swap atomic + pre-restore snapshot created); manual: stop server → list shows local+R2 → restore local round-trips + creates pre-restore-*.db → start, data matches; repeat r2: source; corrupt file rejected without touching app.db. UI: list renders both sources, command string correct.

## Migrations
PR-A2 migration is **optional** — `params` is read fresh and the job defaults missing fields, so code-default `r2_enabled=false` (recommended, mirrors `retain_count` default) avoids a migration. Add explicit `030_*` only if you want a pre-seed. No schema/DDL changes; `schedule_config` unchanged.

## Risks
- New aws-sdk-go-v2 dep in isme + kuery (medioa2 already uses it — platform knows the SDK).
- R2 creds = new isme secrets → `.env` gitignored, prod fly secrets, never echo/commit. Backups carry password hashes + secrets → **private bucket only**.
- Blast radius low: backup job is a leaf scheduled closure; R2 client is a leaf singleton like the medioa client. Only shared touch = kuery new `r2` pkg + minor tag (fine under shared-lib rule).
- DR: a separate backup bucket means R2 loss ≠ host loss (the point of off-host).

## Files (summary)
- **Create:** `kuery/r2/*` (new pkg + tag); `cmd/restore/main.go`; `defineR2Client`/`GetR2Client` in `internal/di/di_service.go`; constant in `internal/constants/di.go`.
- **Edit:** `internal/config/config.go` (R2 struct); `internal/di/di.go` (register); `internal/di/scheduler_jobs.go` (params + body + last_result); `internal/domains/settings/models/database_backup.go`; `internal/domains/settings/usecase/usecase.go` (params/result + `ListRestorePoints`); `internal/domains/settings/handlers/http/{handler.go,route.go}` (restore-points endpoint); `ui/src/components/DatabaseBackupCard.tsx` + `ui/src/apis`; `.env` (local secrets, not committed).
- **Optional:** `db/history/sqlite/030_*` seed.

All work lands via PR (never direct push to main). Each phase (kuery, A1, A2, B) independently verifiable.
