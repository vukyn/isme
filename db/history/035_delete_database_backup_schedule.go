package history

import (
	"context"

	"github.com/uptrace/bun"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"
)

// m035DeleteDatabaseBackupSchedule removes the schedule row of a job that no
// longer exists. Down restores the row as the baseline seeded it.
var m035DeleteDatabaseBackupSchedule = pkgMigrate.Migration{
	Name: "035_delete_database_backup_schedule",
	Up: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(),
			`DELETE FROM schedule_config WHERE job_key = 'database_backup'`)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(),
			`INSERT INTO schedule_config (job_key, enabled, cron, params)
			VALUES ('database_backup', FALSE, '0 3 * * *', '{"retain_count":10}')
			ON CONFLICT (job_key) DO NOTHING`)
		return err
	},
}
