---
name: project-isme-postgres-only
description: isme dropping SQLite (PR2, planned 2026-11-18) — owner decisions, migration-history trap, test-helper choice
metadata:
  type: project
---

Owner decided (2026-11-18) isme becomes Postgres-only in a 3-PR sequence: PR1 atomic invite
Accept + migrations 032-034 (dual-dialect), PR2 remove SQLite + delete the `database_backup`
job (rely on provider backups), PR3 Fiber v3. Tests move to Postgres via testcontainers-go
(Docker on owner machine); the reusable helper is planned for kuery (`bun/pgtest`).

**Why:** prod (fly `vukyn-isme`) already runs DB_DRIVER=postgres; SQLite only survived as
dev default + in-memory tests.

**How to apply:** the trap when collapsing history — on Postgres, migrations 001-029 never
ran (SQLite-only SQL); prod was bootstrapped by `000_baseline`, which stamps EVERY name in
the `Migrations` slice. kuery/bun/migrate tracks by name and `RollbackLast` errors on an
applied name missing from the slice, so keep all 001-034 names (as no-op stubs) and never
rename them. Planned design: baseline first in slice, frozen; legacy names no-op; tail
migrations real. Gate: prod `migrations` table must already hold 000_baseline..034 before
PR2's `up` runs. Owner's local .env points at a Postgres — agents never migrate against it.
Related: [[project-isme-invite-flow]].
