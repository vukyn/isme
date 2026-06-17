package history

import (
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m008CreateRolePermissionsTable = pkgMigrate.Migration{
	Name: "008_create_role_permissions_table",
	Up: func(db *bun.DB) error {
		_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS role_permissions (
				role_id TEXT NOT NULL,
				permission_id INTEGER NOT NULL,
				PRIMARY KEY (role_id, permission_id)
			)
		`)
		return err
	},
	Down: func(db *bun.DB) error {
		_, err := db.Exec(`DROP TABLE IF EXISTS role_permissions`)
		return err
	},
}
