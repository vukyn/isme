package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

// Seed the invitation-cleanup job into schedule_config. It marks pending
// invitations past their expiry as expired. Disabled by default; it has no
// params. Cron 0 6 * * * keeps clear of the other jobs' slots.
var m033SeedInvitationCleanupSchedule = pkgMigrate.Migration{
	Name: "033_seed_invitation_cleanup_schedule",
	Up: func(db bun.IDB) error {
		statement := `
			INSERT OR IGNORE INTO schedule_config (job_key, enabled, cron, params)
			VALUES ('invitation_cleanup', 0, '0 6 * * *', '{}')
		`
		if isPostgres(db) {
			statement = `
				INSERT INTO schedule_config (job_key, enabled, cron, params)
				VALUES ('invitation_cleanup', FALSE, '0 6 * * *', '{}')
				ON CONFLICT (job_key) DO NOTHING
			`
		}
		_, err := db.ExecContext(context.Background(), statement)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `DELETE FROM schedule_config WHERE job_key = 'invitation_cleanup'`)
		return err
	},
}
