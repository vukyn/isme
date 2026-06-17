package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m027SeedActivityCleanupSchedule = pkgMigrate.Migration{
	Name: "027_seed_activity_cleanup_schedule",
	Up: func(db bun.IDB) error {
		// Seed the third scheduled job (activity_cleanup) into the generic
		// schedule_config table — no new table. Disabled by default; a DAY-scale
		// retention (retention_days, default 90) prunes old activity_events.
		// Cron 0 5 * * * avoids the rotation-cleanup job's 0 4 * * * slot.
		_, err := db.ExecContext(context.Background(), `
			INSERT OR IGNORE INTO schedule_config (job_key, enabled, cron, params)
			VALUES ('activity_cleanup', 0, '0 5 * * *', '{"retention_days":90}')
		`)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `DELETE FROM schedule_config WHERE job_key = 'activity_cleanup'`)
		return err
	},
}
