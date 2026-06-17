package history

import (
	"github.com/vukyn/isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
)

// Rotation cleanup scheduler config: a typed single-row table that drives
// the cronjob pruning token_rotation_events older than a configurable
// retention window (the table grows unbounded — one row per refresh).
// The settings:read / settings:update permissions are already seeded by
// migration 022, so this migration only adds the config table.
var m024CreateRotationCleanupConfig = models.Migration{
	Name: "024_create_rotation_cleanup_config",
	Up: func(db *bun.DB) error {
		_, err := db.Exec(`
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
		`)
		if err != nil {
			return err
		}

		// single config row (id=1, disabled by default)
		_, err = db.Exec(`
			INSERT OR IGNORE INTO rotation_cleanup_config (id, enabled, cron, retention_hours)
			VALUES (1, 0, '0 4 * * *', 48)
		`)
		return err
	},
	Down: func(db *bun.DB) error {
		_, err := db.Exec(`DROP TABLE IF EXISTS rotation_cleanup_config`)
		return err
	},
}
