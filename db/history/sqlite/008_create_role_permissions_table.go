package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m008CreateRolePermissionsTable = pkgMigrate.Migration{
	Name: "008_create_role_permissions_table",
	Up: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			CREATE TABLE IF NOT EXISTS role_permissions (
				role_id TEXT NOT NULL,
				permission_id INTEGER NOT NULL,
				PRIMARY KEY (role_id, permission_id)
			)
		`)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `DROP TABLE IF EXISTS role_permissions`)
		return err
	},
}
