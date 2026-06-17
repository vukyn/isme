package history

import (
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m007CreatePermissionsTable = pkgMigrate.Migration{
	Name: "007_create_permissions_table",
	Up: func(db *bun.DB) error {
		_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS permissions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				resource TEXT NOT NULL,
				action TEXT NOT NULL,
				UNIQUE (resource, action)
			)
		`)
		return err
	},
	Down: func(db *bun.DB) error {
		_, err := db.Exec(`DROP TABLE IF EXISTS permissions`)
		return err
	},
}
