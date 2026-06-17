package history

import (
	"github.com/vukyn/isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
)

// Consolidate the two typed scheduled-job config tables
// (session_revoke_config + rotation_cleanup_config) into one generic
// schedule_config keyed by job_key. Shared columns stay typed;
// job-specific knobs move into the params JSON blob and per-run results
// into the last_result JSON blob (replacing last_revoked_count /
// last_cleaned_count). The job-key strings must match the JobKey consts
// in internal/scheduler (asserted by a cross-package test).
var m025ConsolidateScheduleConfig = models.Migration{
	Name: "025_consolidate_schedule_config",
	Up: func(db *bun.DB) error {
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS schedule_config (
				job_key TEXT PRIMARY KEY,
				enabled INTEGER NOT NULL DEFAULT 0,
				cron TEXT NOT NULL,
				params TEXT NOT NULL DEFAULT '{}',
				last_run_at DATETIME,
				last_result TEXT,
				updated_at DATETIME,
				updated_by TEXT
			)
		`); err != nil {
			return err
		}

		// data-migrate the existing session-revoke row (id=1) — empty params,
		// last_revoked_count folds into last_result json {"revoked": N}.
		if _, err := db.Exec(`
			INSERT INTO schedule_config (job_key, enabled, cron, params, last_run_at, last_result, updated_at, updated_by)
			SELECT 'session_revoke', enabled, cron, '{}', last_run_at,
				CASE WHEN last_revoked_count IS NULL THEN NULL ELSE json_object('revoked', last_revoked_count) END,
				updated_at, updated_by
			FROM session_revoke_config WHERE id = 1
		`); err != nil {
			return err
		}

		// data-migrate the existing rotation-cleanup row (id=1) — retention_hours
		// folds into params json, last_cleaned_count into last_result {"cleaned": N}.
		if _, err := db.Exec(`
			INSERT INTO schedule_config (job_key, enabled, cron, params, last_run_at, last_result, updated_at, updated_by)
			SELECT 'rotation_cleanup', enabled, cron, json_object('retention_hours', retention_hours), last_run_at,
				CASE WHEN last_cleaned_count IS NULL THEN NULL ELSE json_object('cleaned', last_cleaned_count) END,
				updated_at, updated_by
			FROM rotation_cleanup_config WHERE id = 1
		`); err != nil {
			return err
		}

		// guard: if a source row was somehow absent, seed the default job rows
		// so a fresh DB still ends with both jobs present (disabled by default).
		if _, err := db.Exec(`
			INSERT OR IGNORE INTO schedule_config (job_key, enabled, cron, params)
			VALUES ('session_revoke', 0, '0 3 * * *', '{}')
		`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			INSERT OR IGNORE INTO schedule_config (job_key, enabled, cron, params)
			VALUES ('rotation_cleanup', 0, '0 4 * * *', json_object('retention_hours', 48))
		`); err != nil {
			return err
		}

		if _, err := db.Exec(`DROP TABLE IF EXISTS session_revoke_config`); err != nil {
			return err
		}
		if _, err := db.Exec(`DROP TABLE IF EXISTS rotation_cleanup_config`); err != nil {
			return err
		}
		return nil
	},
	Down: func(db *bun.DB) error {
		// recreate the two typed tables (mirror of migrations 022 / 024).
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS session_revoke_config (
				id INTEGER PRIMARY KEY,
				enabled INTEGER NOT NULL DEFAULT 0,
				cron TEXT NOT NULL DEFAULT '0 3 * * *',
				last_run_at DATETIME,
				last_revoked_count INTEGER,
				updated_at DATETIME,
				updated_by TEXT
			)
		`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS rotation_cleanup_config (
				id INTEGER PRIMARY KEY,
				enabled INTEGER NOT NULL DEFAULT 0,
				cron TEXT NOT NULL DEFAULT '0 4 * * *',
				retention_hours INTEGER NOT NULL DEFAULT 48,
				last_run_at DATETIME,
				last_cleaned_count INTEGER,
				updated_at DATETIME,
				updated_by TEXT
			)
		`); err != nil {
			return err
		}

		// migrate rows back: extract job-specific knobs out of the JSON blobs.
		if _, err := db.Exec(`
			INSERT OR IGNORE INTO session_revoke_config (id, enabled, cron, last_run_at, last_revoked_count, updated_at, updated_by)
			SELECT 1, enabled, cron, last_run_at, json_extract(last_result, '$.revoked'), updated_at, updated_by
			FROM schedule_config WHERE job_key = 'session_revoke'
		`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			INSERT OR IGNORE INTO rotation_cleanup_config (id, enabled, cron, retention_hours, last_run_at, last_cleaned_count, updated_at, updated_by)
			SELECT 1, enabled, cron,
				COALESCE(json_extract(params, '$.retention_hours'), 48),
				last_run_at, json_extract(last_result, '$.cleaned'), updated_at, updated_by
			FROM schedule_config WHERE job_key = 'rotation_cleanup'
		`); err != nil {
			return err
		}

		// seed defaults if a source row was absent (mirror of 022 / 024).
		if _, err := db.Exec(`
			INSERT OR IGNORE INTO session_revoke_config (id, enabled, cron)
			VALUES (1, 0, '0 3 * * *')
		`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			INSERT OR IGNORE INTO rotation_cleanup_config (id, enabled, cron, retention_hours)
			VALUES (1, 0, '0 4 * * *', 48)
		`); err != nil {
			return err
		}

		_, err := db.Exec(`DROP TABLE IF EXISTS schedule_config`)
		return err
	},
}
