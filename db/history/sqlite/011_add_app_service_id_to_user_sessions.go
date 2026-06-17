package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m011AddAppServiceIDToUserSessions = pkgMigrate.Migration{
	Name: "011_add_app_service_id_to_user_sessions",
	Up: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE user_sessions
			ADD COLUMN app_service_id TEXT NOT NULL DEFAULT ''
		`)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE user_sessions
			DROP COLUMN app_service_id
		`)
		return err
	},
}
