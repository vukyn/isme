package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m005AddIsAdminToUsers = pkgMigrate.Migration{
	Name: "005_add_is_admin_to_users",
	Up: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE users
			ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0
		`)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE users
			DROP COLUMN is_admin
		`)
		return err
	},
}
