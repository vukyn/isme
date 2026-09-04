---
name: postgres-sqlite-gotchas
description: "Postgres vs SQLite dialect traps: ON CONFLICT alias-qualifying (42P01/42702), Go bool needs BOOLEAN (42804), migration id collision (23505), ILIKE is PG-only."
metadata:
  type: project
---

Several memories merged into one topic file to keep the index readable. Each section below is the original entry, unedited.

## pg-upsert-alias-qualify-42p01

In a bun `NewInsert().On("CONFLICT ... DO UPDATE").Set(...)` upsert, bun renders `INSERT INTO <table> AS <alias> ...` (entity `bun:"table:station_listener_stats,alias:sls"`). On Postgres, naming the existing-row value:
- by the **real table name** (`station_listener_stats.col`) → `42P01 invalid reference to FROM-clause entry` (only alias `sls` is in scope);
- by the **bare column** (`col`) → `42702 column reference is ambiguous` (both the aliased target and `EXCLUDED` have it).
So PG MUST use the **alias**: `GREATEST(sls.col, EXCLUDED.col)`. **SQLite tolerates** (and historically used) the real table name. The bug only surfaces on the Postgres cutover. Verified on throwaway docker PG.

Correct dialect-branched form (LHS is the bare assignment target = fine on both; only the RHS existing-value qualifier + scalar fn differ):
- SQLite: `peak_concurrent = MAX(station_listener_stats.peak_concurrent, EXCLUDED.peak_concurrent)` (real name; scalar `MAX(a,b)`)
- PG:     `peak_concurrent = GREATEST(sls.peak_concurrent, EXCLUDED.peak_concurrent)` (alias; PG has no scalar MAX → GREATEST)

**Why:** dual-DB SQLite+Postgres ([[dual-db-sqlite-postgres-plan]]) hides dialect bugs until the PG cutover — like [[postgres-sqlite-gotchas]]. Hit in rainy `station_listener_stat/repository/repository.go` UpsertDailyPeak (listener-stats peak sampler [[rainy-station-and-listener]]), 2026-06-23.

**How to apply:** never qualify the conflict-target table by its real name in a bun upsert SET expression — use the bare column (or the bun alias). Grep upsert `.Set(` fragments for `<tablename>.` qualifiers when porting a service to Postgres.

## pg-bool-column-and-migrate-id-collision

isme prod runs on **Neon Postgres** (`.env` `DB_DSN=postgresql://neondb_owner:...neon.tech/isme`), not SQLite — `DB_DSN` overrides `DB_HOST/PORT` in `kuery/bun/db.Open`. ⚠️ running `go run db/migrate.go postgres ...` locally hits LIVE PROD unless you blank `DB_DSN` and set `DB_HOST/PORT` to a throwaway container.

**Bug 1 — bool→INTEGER type clash (SQLSTATE 42804).** bun's pgdialect serializes a Go `bool` field as the literal `TRUE/FALSE`. SQLite stores bool as INTEGER and is loose, so it works there; Postgres is strict → `column "x" is of type integer but expression is of type boolean`. Any column whose Go entity field is `bool` MUST be declared `BOOLEAN` on the PG branch (not INTEGER). isme had 3: `users.is_verified`, `roles.is_system`, `schedule_config.enabled` — only `enabled` was hit first (session-revoke settings toggle, 2026-06-22). Fixed: baseline.go PG branch → BOOLEAN + seed literals `0/1`→`FALSE/TRUE`; new migration `030_fix_bool_columns_pg` ALTERs existing prod (`TYPE BOOLEAN USING (col<>0)`, PG-only, SQLite no-op). **When auditing another service for PG: grep entity `bool` fields, confirm the PG column is BOOLEAN.** rainy hit the SAME bug 2026-06-22 (worker 42883 `integer = boolean` on `GetActivatable` `enabled = ?`): `schedule_config.enabled` + `station_schedules.enabled` were INTEGER on both dialect branches (migrations 026/037); other rainy bool cols (success/truncated/public/use_youtube_cover) were already BOOLEAN. Fixed: baseline PG → BOOLEAN + seed FALSE, migration `039_fix_enabled_bool_columns_pg` (defaults FALSE/TRUE per col), kuery bump v1.33→v1.36.0. rainy PR #203 merged + 039 applied to prod rainy Neon + local PG. **Sibling gotcha (reverse direction):** raw query predicates binding an INTEGER literal against a BOOLEAN column also fail on PG (`boolean = integer` 42883) — rainy `ListLivePublic` had `Where("public = ?", 1)` (sampler logged it every tick); fix = bind Go `true` so bun renders per-dialect (PR #204). grep `Where("...= ?", 1|0)` for the pattern.

**Bug 2 — migrate runner id collision after baseline.** `kuery/bun/migrate.Run` computed next id from `executedMigrations[len-1].ID` on an **unordered** SELECT. Postgres doesn't guarantee row order, so after a baseline stamps ids 1..N, the FIRST new incremental migration picks a non-max id → `duplicate key ... migrations_pkey` (23505). Fixed to `max(id)` over all rows. Any service using baseline+incremental on PG would hit this on its next migration.

**How to apply:** both fixes ship in kuery (Err funnel + runner) and isme. Deploy order: tag kuery (keep 5 newest) → `go get` in isme → `migrate up` on prod (applies m030) → redeploy. See [[dual-db-sqlite-postgres-plan]], [[deploy-env-optional]], [[never-rm-dev-db-in-smoke]], [[gopls-stale-diagnostics-multirepo]].

Also fixed in same change: `kuery/http/fiber.Err` leaked raw 5xx `pkgErr.Error` messages (e.g. raw DB error) to clients — now logs detail server-side + returns generic `"internal server error"` for any status ≥500 (matches what the `default` branch already did). Platform-wide for all kuery consumers.

## sqlite-no-ilike-use-lower-like

isme/rainy run on SQLite (Bun). `ILIKE` is Postgres-only — using it raises `SQL logic error: near "ILIKE": syntax error (1)` → HTTP 500 at runtime (build/tests pass; raw SQL string not type-checked). Hit on isme `GET /api/v1/users?query=` (user list search), fixed in PR #41 by switching `name ILIKE ?` → `LOWER(name) LIKE LOWER(?)`.

**Why:** the shared clean-arch template is copied across services and any `repository.go` search/filter that hand-writes `ILIKE` will crash only when the search param is non-empty — easy to miss until a query reaches it.

**How to apply:** for case-insensitive matching in any SQLite-backed service (isme, rainy) use `LOWER(col) LIKE LOWER(?)` with `%term%`; never `ILIKE`. When reviewing/adding a list endpoint's search `Where`, grep the service for `ILIKE`. medioa2/memz use MongoDB so this doesn't apply there.
