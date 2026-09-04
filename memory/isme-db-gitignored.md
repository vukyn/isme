---
name: isme-db-gitignored
description: isme/db/app.db is now untracked + gitignored — never commit it; recreate locally via migrate-up
metadata: 
  node_type: memory
  type: project
---

As of 2026-06-08 (PR #26), `isme/db/app.db` is **untracked and gitignored** (`db/*.db` in isme/.gitignore). It was previously committed and leaked real secrets — GitGuardian flagged it: `app_services.app_secret` (AES-encrypted, low risk — AES_SECRET not in repo) and `user_sessions.refresh_token` (64 rows). Both were rotated: app_service secrets regenerated, medioa2/.env `AUTH_APP_SECRET` updated to match (app_code `my-lab-medioa2`), all sessions wiped.

**Why:** local SQLite holds runtime secrets; schema is reproducible from migrations, the binary is not source.

**How to apply:**
- Never re-add `db/app.db` to git. Don't expect it in a fresh clone.
- Recreate an empty local db with `go run db/migrate.go sqlite up` (NOT `DB=app` — root CLAUDE.md is stale on this; migrate.go only accepts `sqlite`).
- Switching/pulling a branch whose merge deletes a now-untracked db can wipe the working copy — back up `db/app.db` before branch ops.
- `medioa2/.env` is **gitignored + untracked, never committed** (verified `git log --all -- .env` empty) — the rotated app_secret stays local, no leak. Root CLAUDE.md's "committed .env" note is stale/wrong. Still never echo it (live R2/Mongo creds). See [[commits-always-via-pr]].
