package history

import (
	"github.com/vukyn/isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
)

var m027SeedActivityCleanupSchedule = models.Migration{
	Name: "027_seed_activity_cleanup_schedule",
	Up: func(db *bun.DB) error {
		// Seed the third scheduled job (activity_cleanup) into the generic
		// schedule_config table — no new table. Disabled by default; a DAY-scale
		// retention (retention_days, default 90) prunes old activity_events.
		// Cron 0 5 * * * avoids the rotation-cleanup job's 0 4 * * * slot.
		_, err := db.Exec(`
			INSERT OR IGNORE INTO schedule_config (job_key, enabled, cron, params)
			VALUES ('activity_cleanup', 0, '0 5 * * *', '{"retention_days":90}')
		`)
		return err
	},
	Down: func(db *bun.DB) error {
		_, err := db.Exec(`DELETE FROM schedule_config WHERE job_key = 'activity_cleanup'`)
		return err
	},
}
