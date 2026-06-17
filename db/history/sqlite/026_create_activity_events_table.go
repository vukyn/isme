package history

import (
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m026CreateActivityEventsTable = pkgMigrate.Migration{
	Name: "026_create_activity_events_table",
	Up: func(db *bun.DB) error {
		// activity_events is an append-only audit log powering the Welcome
		// "Recent activity" feed. NOTE: this table grows unbounded — one row
		// per recorded user action (sign_in/out, password change, invitation).
		// Retention/pruning is deliberately deferred (no scheduler job yet);
		// add a prune job before this becomes a volume concern.
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS activity_events (
				id TEXT PRIMARY KEY NOT NULL,
				user_id TEXT NOT NULL,
				type TEXT NOT NULL,
				meta TEXT NOT NULL DEFAULT '{}',
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`); err != nil {
			return err
		}
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_activity_events_user_created ON activity_events (user_id, created_at)`); err != nil {
			return err
		}
		return nil
	},
	Down: func(db *bun.DB) error {
		if _, err := db.Exec(`DROP INDEX IF EXISTS idx_activity_events_user_created`); err != nil {
			return err
		}
		_, err := db.Exec(`DROP TABLE IF EXISTS activity_events`)
		return err
	},
}
