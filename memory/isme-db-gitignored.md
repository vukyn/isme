---
name: isme-db-gitignored
description: isme/db/app.db + db/backups/ are LEGACY SQLite files — isme no longer uses them; stay gitignored, never commit, never delete them for the owner
metadata: 
  node_type: memory
  type: project
---

isme has been **Postgres-only since 2026-11-18**. `db/app.db` and `db/backups/` are leftovers from the SQLite era: nothing reads or writes them any more, but they may still sit in a working copy and they hold real secrets.

History: as of 2026-06-08 (PR #26) `db/app.db` was untracked and gitignored. It had been committed before and leaked real secrets — GitGuardian flagged `app_services.app_secret` (AES-encrypted, low risk — AES_SECRET not in repo) and `user_sessions.refresh_token` (64 rows). Both were rotated: app_service secrets regenerated, medioa2's `AUTH_APP_SECRET` updated to match, all sessions wiped. `db/backups/` held `VACUUM INTO` snapshots from the deleted `database_backup` job — the same secrets again.

**Why:** the files are binary copies of runtime secrets; the ignore lines are the only thing between them and a commit.

**How to apply:**
- Keep the `db/*.db` and `db/backups/` lines in `.gitignore` / `.dockerignore` (commented "legacy SQLite files — never commit"). Do not "clean them up" just because the code no longer needs them.
- Never commit, print or ship these files, and never delete them on the owner's behalf — whether to discard them is the owner's call.
- A local database is Postgres now: `make db-up` → `make migrate-up` → `make run`.
- `medioa2/.env` is **gitignored + untracked, never committed** (verified `git log --all -- .env` empty) — still never echo it (live R2/Mongo creds). See [[commits-always-via-pr]].
