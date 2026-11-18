package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

// Seed the unverified-user-cleanup job into schedule_config. It soft-deletes
// accounts that never finished sign-up once they are older than min_age_days.
// Disabled by default because it deletes accounts. Cron 0 7 * * *.
var m034SeedUnverifiedUserCleanupSchedule = pkgMigrate.Migration{
	Name: "034_seed_unverified_user_cleanup_schedule",
	Up: func(db bun.IDB) error {
		statement := `
			INSERT OR IGNORE INTO schedule_config (job_key, enabled, cron, params)
			VALUES ('unverified_user_cleanup', 0, '0 7 * * *', '{"min_age_days":30}')
		`
		if isPostgres(db) {
			statement = `
				INSERT INTO schedule_config (job_key, enabled, cron, params)
				VALUES ('unverified_user_cleanup', FALSE, '0 7 * * *', '{"min_age_days":30}')
				ON CONFLICT (job_key) DO NOTHING
			`
		}
		_, err := db.ExecContext(context.Background(), statement)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `DELETE FROM schedule_config WHERE job_key = 'unverified_user_cleanup'`)
		return err
	},
}
