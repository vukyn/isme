package history

import (
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m011AddAppServiceIDToUserSessions = pkgMigrate.Migration{
	Name: "011_add_app_service_id_to_user_sessions",
	Up: func(db *bun.DB) error {
		_, err := db.Exec(`
			ALTER TABLE user_sessions
			ADD COLUMN app_service_id TEXT NOT NULL DEFAULT ''
		`)
		return err
	},
	Down: func(db *bun.DB) error {
		_, err := db.Exec(`
			ALTER TABLE user_sessions
			DROP COLUMN app_service_id
		`)
		return err
	},
}
