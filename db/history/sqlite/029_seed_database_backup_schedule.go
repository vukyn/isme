package history

import (
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m029SeedDatabaseBackupSchedule = pkgMigrate.Migration{
	Name: "029_seed_database_backup_schedule",
	Up: func(db *bun.DB) error {
		// Seed the fourth scheduled job (database_backup) into the generic
		// schedule_config table — no new table. Disabled by default; a COUNT-scale
		// retain (retain_count, default 10) prunes old backup files.
		// Cron 0 3 * * * avoids the rotation-cleanup job's 0 4 * * * slot and the
		// activity-cleanup job's 0 5 * * * slot.
		_, err := db.Exec(`
			INSERT OR IGNORE INTO schedule_config (job_key, enabled, cron, params)
			VALUES ('database_backup', 0, '0 3 * * *', '{"retain_count":10}')
		`)
		return err
	},
	Down: func(db *bun.DB) error {
		_, err := db.Exec(`DELETE FROM schedule_config WHERE job_key = 'database_backup'`)
		return err
	},
}
